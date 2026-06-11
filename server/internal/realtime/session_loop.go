package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/logging"
	"github.com/mmandrille_meli/arpg-dev/server/internal/replay"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

type sessionLoop struct {
	hub  *Hub
	sess store.Session
	sim  *game.Sim
	log  *slog.Logger

	mu       sync.Mutex
	clients  map[string]*loopClient
	buffer   map[uint64][]game.Input
	seen     map[string]bool
	received map[string]time.Time
	seq      int64

	done      chan struct{}
	closeOnce sync.Once
}

type loopClient struct {
	loop     *sessionLoop
	conn     *websocket.Conn
	key      string
	member   store.SessionMember
	playerID uint64
	sendCh   chan outEnvelope
	done     chan struct{}
	once     sync.Once
}

func newSessionLoop(ctx context.Context, h *Hub, sess store.Session) (*sessionLoop, error) {
	sim, meta, err := buildSessionSim(ctx, h, sess)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	seq := int64(0)
	if meta != nil {
		for id := range meta.SeenMessageIDs {
			seen[id] = true
		}
		seq = meta.NextSequence
	}
	return &sessionLoop{
		hub:      h,
		sess:     sess,
		sim:      sim,
		log:      logging.Component(h.log, "realtime").With("session_id", sess.ID),
		clients:  make(map[string]*loopClient),
		buffer:   make(map[uint64][]game.Input),
		seen:     seen,
		received: make(map[string]time.Time),
		seq:      seq,
		done:     make(chan struct{}),
	}, nil
}

func buildSessionSim(ctx context.Context, h *Hub, sess store.Session) (*game.Sim, *replay.ResumeMetadata, error) {
	storedInputs, err := h.store.ListInputs(ctx, sess.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("list inputs: %w", err)
	}
	if len(storedInputs) > 0 {
		recon, err := replay.Reconstruct(ctx, h.store, h.rules, sess.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("reconstruct session: %w", err)
		}
		return recon.Sim, &recon.Metadata, nil
	}

	members, err := h.store.ListSessionMembers(ctx, sess.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("list members: %w", err)
	}
	if len(members) == 0 {
		members = []store.SessionMember{{
			SessionID:   sess.ID,
			AccountID:   sess.AccountID,
			CharacterID: sess.CharacterID,
			Role:        store.SessionMemberHost,
			Status:      store.SessionMemberActive,
		}}
	}
	sort.Slice(members, func(i, j int) bool {
		if members[i].Role != members[j].Role {
			return members[i].Role == store.SessionMemberHost
		}
		if members[i].JoinedTick != members[j].JoinedTick {
			return members[i].JoinedTick < members[j].JoinedTick
		}
		if members[i].AccountID != members[j].AccountID {
			return members[i].AccountID < members[j].AccountID
		}
		return members[i].CharacterID < members[j].CharacterID
	})

	host := members[0]
	for _, member := range members {
		if member.Role == store.SessionMemberHost {
			host = member
			break
		}
	}
	worldID := sess.WorldID
	if worldID == "" {
		worldID = game.DefaultWorldID
	}
	hostStart, err := h.store.LoadSessionStartSnapshotForMember(ctx, sess.ID, host.AccountID, host.CharacterID)
	if err != nil {
		return nil, nil, fmt.Errorf("load host start snapshot: %w", err)
	}
	hostProgression, err := h.progressionStateForMember(ctx, host, hostStart.Progression)
	if err != nil {
		return nil, nil, err
	}
	sim, err := game.NewSimWithWorldProgression(sess.ID, sess.Seed, h.rules, worldID, hostProgression)
	if err != nil {
		return nil, nil, err
	}
	hostPlayerID := sim.DefaultPlayerID()
	sim.SetPlayerMetadata(hostPlayerID, host.AccountID, host.CharacterID, "Hero", store.SessionMemberHost)
	sim.LoadInventoryForPlayer(hostPlayerID, persistedItems(hostStart.Items))
	sim.LoadHotbarForPlayer(hostPlayerID, persistedHotbar(hostStart.Hotbar))
	sim.LoadSkillBindingsForPlayer(hostPlayerID, persistedSkillBindings(hostStart.SkillBinds))
	sim.LoadDiscoveredTeleportersForPlayer(hostPlayerID, waypointLevels(hostStart.Waypoints))
	sim.LoadShopStockForPlayer(hostPlayerID, persistedShopStock(hostStart.ShopStock))
	sim.LoadAccountStashForPlayer(hostPlayerID, persistedStashItems(hostStart.StashItems), hostStart.StashGold.Gold, 0)
	if err := h.store.SetSessionMemberPlayer(ctx, sess.ID, host.AccountID, host.CharacterID, idStr(hostPlayerID), 0); err != nil && err != store.ErrNotFound {
		return nil, nil, err
	}

	for _, member := range members {
		if member.AccountID == host.AccountID && member.CharacterID == host.CharacterID {
			continue
		}
		start, err := h.store.LoadSessionStartSnapshotForMember(ctx, sess.ID, member.AccountID, member.CharacterID)
		if err != nil {
			return nil, nil, fmt.Errorf("load guest start snapshot: %w", err)
		}
		memberProgression, err := h.progressionStateForMember(ctx, member, start.Progression)
		if err != nil {
			return nil, nil, err
		}
		playerID, err := sim.AddGuestPlayer(member.AccountID, member.CharacterID, "Guest", memberProgression)
		if err != nil {
			return nil, nil, err
		}
		sim.LoadInventoryForPlayer(playerID, persistedItems(start.Items))
		sim.LoadHotbarForPlayer(playerID, persistedHotbar(start.Hotbar))
		sim.LoadSkillBindingsForPlayer(playerID, persistedSkillBindings(start.SkillBinds))
		sim.LoadDiscoveredTeleportersForPlayer(playerID, waypointLevels(start.Waypoints))
		sim.LoadShopStockForPlayer(playerID, persistedShopStock(start.ShopStock))
		sim.LoadAccountStashForPlayer(playerID, persistedStashItems(start.StashItems), start.StashGold.Gold, 0)
		if err := h.store.SetSessionMemberPlayer(ctx, sess.ID, member.AccountID, member.CharacterID, idStr(playerID), 0); err != nil && err != store.ErrNotFound {
			return nil, nil, err
		}
	}
	return sim, nil, nil
}

func (h *Hub) progressionStateForMember(ctx context.Context, member store.SessionMember, progression *store.CharacterProgression) (game.CharacterProgressionState, error) {
	if progression == nil {
		return progressionStateFromStore(h.rules, progression), nil
	}
	character, err := h.store.GetCharacter(ctx, member.CharacterID)
	if err != nil {
		return game.CharacterProgressionState{}, fmt.Errorf("load character class: %w", err)
	}
	if character.CharacterClass == "" || character.CharacterClass == progression.CharacterClass {
		return progressionStateFromStore(h.rules, progression), nil
	}
	updated := *progression
	updated.CharacterClass = character.CharacterClass
	return progressionStateFromStore(h.rules, &updated), nil
}

func (l *sessionLoop) start() {
	go l.tickLoop()
}

func (l *sessionLoop) stop() {
	l.closeOnce.Do(func() {
		close(l.done)
	})
}

func (l *sessionLoop) hasConnectedMember(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.clients[key] != nil
}

func (l *sessionLoop) attach(ctx context.Context, conn *websocket.Conn, member store.SessionMember) {
	playerID := l.playerIDForMember(ctx, member)
	client := &loopClient{
		loop:     l,
		conn:     conn,
		key:      memberKey(member),
		member:   member,
		playerID: playerID,
		sendCh:   make(chan outEnvelope, sendQueueSize),
		done:     make(chan struct{}),
	}

	l.mu.Lock()
	l.clients[client.key] = client
	isCoopMember := isCoopSession(l.sess) ||
		member.AccountID != l.sess.AccountID ||
		member.CharacterID != l.sess.CharacterID
	currentLevel, _ := l.sim.PlayerCurrentLevel(playerID)
	if (isCoopMember || playerID != l.sim.DefaultPlayerID()) && (!member.Connected || member.CurrentLevel != 0 || currentLevel != 0 || !l.sim.PlayerConnected(playerID)) {
		if err := l.sim.RespawnPlayerInTown(playerID); err != nil {
			l.log.Error("respawn reconnecting player", "player_id", playerID, "error", err)
		}
	}
	if level, ok := l.sim.PlayerCurrentLevel(playerID); ok {
		_ = l.hub.store.SetSessionMemberConnected(context.Background(), member.SessionID, member.AccountID, member.CharacterID, idStr(playerID), level, int64(l.sim.CurrentTick()))
		l.sim.SetPlayerConnected(playerID, true)
	}
	l.mu.Unlock()

	l.hub.metrics.WSConnections.Inc()
	go client.writeLoop()
	go client.readLoop()
	client.enqueue(l.snapshotEnvelope(playerID))
	l.broadcastSnapshots()
}

func (l *sessionLoop) playerIDForMember(ctx context.Context, member store.SessionMember) uint64 {
	if id, ok := game.ParseEntityID(member.PlayerEntityID); ok && id != 0 {
		l.mu.Lock()
		_, exists := l.sim.PlayerCurrentLevel(id)
		l.mu.Unlock()
		if exists {
			return id
		}
	}
	l.mu.Lock()
	if playerID, ok := l.sim.PlayerIDForCharacter(member.CharacterID); ok {
		l.mu.Unlock()
		return playerID
	}
	l.mu.Unlock()

	start, err := l.hub.store.LoadSessionStartSnapshotForMember(ctx, member.SessionID, member.AccountID, member.CharacterID)
	if err != nil {
		l.log.Error("load late-joined member start snapshot", "account_id", member.AccountID, "character_id", member.CharacterID, "error", err)
		return l.sim.DefaultPlayerID()
	}

	l.mu.Lock()
	if playerID, ok := l.sim.PlayerIDForCharacter(member.CharacterID); ok {
		l.mu.Unlock()
		return playerID
	}
	playerID, err := l.sim.AddGuestPlayer(member.AccountID, member.CharacterID, displayNameForMember(member), progressionStateFromStore(l.hub.rules, start.Progression))
	if err != nil {
		l.mu.Unlock()
		l.log.Error("add late-joined guest player", "account_id", member.AccountID, "character_id", member.CharacterID, "error", err)
		return l.sim.DefaultPlayerID()
	}
	l.sim.LoadInventoryForPlayer(playerID, persistedItems(start.Items))
	l.sim.LoadHotbarForPlayer(playerID, persistedHotbar(start.Hotbar))
	l.sim.LoadSkillBindingsForPlayer(playerID, persistedSkillBindings(start.SkillBinds))
	l.sim.LoadDiscoveredTeleportersForPlayer(playerID, waypointLevels(start.Waypoints))
	l.sim.LoadShopStockForPlayer(playerID, persistedShopStock(start.ShopStock))
	l.sim.LoadAccountStashForPlayer(playerID, persistedStashItems(start.StashItems), start.StashGold.Gold, 0)
	l.mu.Unlock()
	if err := l.hub.store.SetSessionMemberPlayer(context.Background(), member.SessionID, member.AccountID, member.CharacterID, idStr(playerID), 0); err != nil && err != store.ErrNotFound {
		l.log.Error("set late-joined member player", "account_id", member.AccountID, "character_id", member.CharacterID, "player_id", playerID, "error", err)
	}
	return playerID
}

func (l *sessionLoop) detach(client *loopClient) {
	client.close()
	l.mu.Lock()
	if l.clients[client.key] != client {
		l.mu.Unlock()
		return
	}
	delete(l.clients, client.key)
	level, _ := l.sim.PlayerCurrentLevel(client.playerID)
	if isCoopSession(l.sess) {
		l.sim.RemovePlayerEntity(client.playerID)
	}
	_ = l.hub.store.SetSessionMemberDisconnected(context.Background(), l.sess.ID, client.member.AccountID, client.member.CharacterID, level, int64(l.sim.CurrentTick()))
	remaining := len(l.clients)
	clients := l.clientsForLevelLocked(level)
	tick := l.sim.CurrentTick()
	l.mu.Unlock()

	if isCoopSession(l.sess) && l.sess.Listed && remaining == 0 {
		if _, err := l.hub.store.EndListedSessionIfNoConnected(context.Background(), l.sess.ID); err != nil {
			l.log.Error("end empty listed session", "session_id", l.sess.ID, "error", err)
		}
	}

	if isCoopSession(l.sess) {
		remove := outEnvelope{
			Type:      typeStateDelta,
			MessageID: ids.New("msg"),
			SessionID: l.sess.ID,
			Tick:      tick,
			Payload: stateDeltaPayload{
				ServerTick: tick,
				Level:      level,
				Changes:    []game.Change{{Op: game.OpEntityRemove, EntityID: idStr(client.playerID)}},
				Events:     []game.Event{},
			},
		}
		for _, other := range clients {
			other.enqueue(remove)
		}
	}
	if remaining == 0 {
		l.stop()
		l.hub.removeLoop(l.sess.ID, l)
	}
}

func (l *sessionLoop) snapshotEnvelope(playerID uint64) outEnvelope {
	snap := l.sim.SnapshotForPlayer(playerID)
	return outEnvelope{
		Type:      typeSnapshot,
		MessageID: ids.New("msg"),
		SessionID: l.sess.ID,
		Tick:      snap.ServerTick,
		Payload:   snap,
	}
}

func (l *sessionLoop) broadcastSnapshots() {
	l.mu.Lock()
	clients := make([]*loopClient, 0, len(l.clients))
	membersByPlayerID := make(map[uint64]store.SessionMember, len(l.clients))
	for _, client := range l.clients {
		clients = append(clients, client)
		membersByPlayerID[client.playerID] = client.member
	}
	l.mu.Unlock()
	for _, client := range clients {
		client.enqueue(l.snapshotEnvelope(client.playerID))
	}
}

func (c *loopClient) readLoop() {
	defer c.loop.detach(c)
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		c.loop.handleClientMessage(c, data)
	}
}

func (c *loopClient) writeLoop() {
	defer c.loop.detach(c)
	for {
		select {
		case <-c.done:
			return
		case env := <-c.sendCh:
			if err := c.conn.WriteJSON(env); err != nil {
				return
			}
		}
	}
}

func (c *loopClient) enqueue(env outEnvelope) {
	select {
	case <-c.done:
	case c.sendCh <- env:
	default:
		c.close()
	}
}

func (c *loopClient) close() {
	c.once.Do(func() {
		close(c.done)
		_ = c.conn.Close()
		c.loop.hub.metrics.WSConnections.Dec()
	})
}

func (l *sessionLoop) handleClientMessage(client *loopClient, data []byte) {
	var env inEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		client.enqueue(l.errorEnvelope("bad_message", "malformed JSON envelope"))
		return
	}
	if env.Type == "" || env.MessageID == "" {
		client.enqueue(l.errorEnvelope("bad_message", "envelope missing type or message_id"))
		return
	}
	if env.SessionID != "" && env.SessionID != l.sess.ID {
		client.enqueue(l.errorEnvelope("bad_session", "session_id does not match this connection"))
		return
	}
	if env.Type == typeClientReady {
		client.enqueue(l.snapshotEnvelope(client.playerID))
		client.enqueue(l.acceptedEnvelope(env.MessageID, l.currentTick(), env.CorrelationID))
		return
	}
	if !isClientIntent(env.Type) {
		client.enqueue(l.errorEnvelope("bad_message", "unknown message type: "+env.Type))
		return
	}
	in, ok := decodeInput(env)
	if !ok {
		client.enqueue(l.rejectedEnvelope(env.MessageID, "invalid_payload", env.CorrelationID))
		return
	}
	in.ActorPlayerID = client.playerID

	l.mu.Lock()
	if l.seen[env.MessageID] {
		l.mu.Unlock()
		client.enqueue(l.rejectedEnvelope(env.MessageID, "duplicate", env.CorrelationID))
		return
	}
	l.seen[env.MessageID] = true
	cur := l.sim.CurrentTick()
	tick := env.Tick
	if tick < cur {
		tick = cur
	}
	in.Sequence = l.seq
	l.seq++
	l.buffer[tick] = append(l.buffer[tick], in)
	l.received[env.MessageID] = time.Now()
	rec := store.SessionInput{
		ID:                  ids.New("inp"),
		SessionID:           l.sess.ID,
		Tick:                int64(tick),
		Sequence:            in.Sequence,
		MessageID:           env.MessageID,
		CorrelationID:       env.CorrelationID,
		ActorAccountID:      client.member.AccountID,
		ActorCharacterID:    client.member.CharacterID,
		ActorPlayerEntityID: idStr(client.playerID),
		Payload:             json.RawMessage(data),
	}
	l.mu.Unlock()

	if err := l.hub.store.AppendInput(context.Background(), rec); err != nil {
		l.hub.metrics.PersistenceErrors.Inc()
		l.log.Error("persist input", "error", err)
	}
}

func (l *sessionLoop) tickLoop() {
	ticker := time.NewTicker(time.Second / tickHz)
	defer ticker.Stop()
	for {
		select {
		case <-l.done:
			return
		case <-ticker.C:
			l.doTick()
		}
	}
}

func (l *sessionLoop) doTick() {
	start := time.Now()
	l.mu.Lock()
	tick := l.sim.CurrentTick()
	inputs := l.buffer[tick]
	inputTypes := make(map[string]string, len(inputs))
	for _, in := range inputs {
		inputTypes[in.MessageID] = in.Type
	}
	delete(l.buffer, tick)
	sortInputs(inputs)
	results := l.sim.TickResults(inputs)
	latencies := []time.Duration{}
	for _, res := range results {
		for _, ack := range res.Acks {
			if recv, ok := l.received[ack.MessageID]; ok {
				latencies = append(latencies, time.Since(recv))
				delete(l.received, ack.MessageID)
			}
		}
	}
	clients := make([]*loopClient, 0, len(l.clients))
	membersByPlayerID := make(map[uint64]store.SessionMember, len(l.clients))
	for _, client := range l.clients {
		clients = append(clients, client)
		membersByPlayerID[client.playerID] = client.member
	}
	l.mu.Unlock()

	l.hub.metrics.TickDuration.Observe(time.Since(start).Seconds())
	for _, latency := range latencies {
		l.hub.metrics.MessageLatency.Observe(latency.Seconds())
	}
	eventSequence := int64(0)
	for _, res := range results {
		eventSequence = l.persistTick(res, membersByPlayerID, eventSequence)
		l.fanoutResult(res, clients, inputTypes)
	}
}

func (l *sessionLoop) fanoutResult(res game.TickResult, clients []*loopClient, inputTypes map[string]string) {
	for _, client := range clients {
		level, ok := l.sim.PlayerCurrentLevel(client.playerID)
		if !ok {
			continue
		}
		for _, ack := range res.Acks {
			if res.ActorPlayerID == client.playerID {
				client.enqueue(l.acceptedEnvelope(ack.MessageID, res.Tick, ""))
				if isInventoryIntentType(inputTypes[ack.MessageID]) {
					_ = inputTypes
				}
			}
		}
		for _, rej := range res.Rejects {
			if res.ActorPlayerID == client.playerID {
				client.enqueue(l.rejectedEnvelope(rej.MessageID, rej.Reason, ""))
			}
		}
		events := filterEventsForClient(res.Events, res.ActorPlayerID, client.playerID)
		if level != res.Level {
			if len(events) == 0 {
				continue
			}
			client.enqueue(outEnvelope{
				Type:      typeStateDelta,
				MessageID: ids.New("msg"),
				SessionID: l.sess.ID,
				Tick:      res.Tick,
				Payload: stateDeltaPayload{
					ServerTick: res.Tick,
					Level:      level,
					Changes:    []game.Change{},
					Events:     events,
				},
			})
			continue
		}
		changes := filterChangesForClient(res.Changes, res.ActorPlayerID, client.playerID)
		if len(changes) == 0 && len(events) == 0 {
			continue
		}
		client.enqueue(outEnvelope{
			Type:      typeStateDelta,
			MessageID: ids.New("msg"),
			SessionID: l.sess.ID,
			Tick:      res.Tick,
			Payload: stateDeltaPayload{
				ServerTick: res.Tick,
				Level:      res.Level,
				Changes:    changes,
				Events:     events,
			},
		})
	}
}

func filterChangesForClient(changes []game.Change, actorPlayerID, clientPlayerID uint64) []game.Change {
	out := make([]game.Change, 0, len(changes))
	for _, change := range changes {
		switch change.Op {
		case game.OpInventoryAdd, game.OpInventoryUpdate, game.OpInventoryRemove,
			game.OpEquippedUpdate, game.OpHotbarUpdate, game.OpTeleporterDiscoveryUpdate,
			game.OpGoldUpdate, game.OpCharacterProgressionUpdate, game.OpSkillProgressionUpdate,
			game.OpShopStockReplace, game.OpShopStockAvailability,
			game.OpStashItemAdd, game.OpStashItemRemove, game.OpStashGoldUpdate:
			ownerPlayerID := actorPlayerID
			if change.OwnerPlayerID != 0 {
				ownerPlayerID = change.OwnerPlayerID
			}
			if ownerPlayerID == 0 || ownerPlayerID != clientPlayerID {
				continue
			}
		}
		out = append(out, change)
	}
	return out
}

func filterEventsForClient(events []game.Event, actorPlayerID, clientPlayerID uint64) []game.Event {
	out := make([]game.Event, 0, len(events))
	for _, event := range events {
		switch event.EventType {
		case "level_changed", "shop_opened", "shop_purchase", "shop_sale", "stash_opened", "stash_item_deposited", "stash_item_withdrawn", "stash_gold_deposited", "stash_gold_withdrawn":
			if actorPlayerID != clientPlayerID {
				continue
			}
		case "experience_gained", "character_leveled", "skill_point_gained", "gold_picked_up":
			ownerPlayerID, ok := eventEntityPlayerID(event)
			if !ok || ownerPlayerID != clientPlayerID {
				continue
			}
		}
		out = append(out, event)
	}
	return out
}

func (l *sessionLoop) persistTick(res game.TickResult, membersByPlayerID map[uint64]store.SessionMember, eventSequence int64) int64 {
	ctx := context.Background()
	member := store.SessionMember{
		AccountID:   l.sess.AccountID,
		CharacterID: l.sess.CharacterID,
	}
	if res.ActorPlayerID != 0 {
		if actorMember, ok := membersByPlayerID[res.ActorPlayerID]; ok {
			member = actorMember
		}
	}

	for _, ev := range res.Events {
		payload, _ := json.Marshal(ev)
		if err := l.hub.store.AppendEvent(ctx, store.SessionEvent{
			ID:            ids.New("evt"),
			SessionID:     l.sess.ID,
			Tick:          int64(res.Tick),
			Sequence:      eventSequence,
			EventType:     ev.EventType,
			CorrelationID: ev.CorrelationID,
			Payload:       payload,
		}); err != nil {
			l.hub.metrics.PersistenceErrors.Inc()
			l.log.Error("persist event", "error", err)
		}
		if ev.EventType == "player_killed" {
			if member, ok := killedEventMember(ev, membersByPlayerID); ok {
				if err := l.hub.store.MarkCharacterDead(ctx, member.AccountID, member.CharacterID); err != nil && !errors.Is(err, store.ErrNotFound) {
					l.hub.metrics.PersistenceErrors.Inc()
					l.log.Error("persist character death", "account_id", member.AccountID, "character_id", member.CharacterID, "error", err)
				}
			}
		}
		switch ev.EventType {
		case "stash_item_deposited":
			if ev.ItemInstanceID == "" || ev.StashItemID == "" {
				break
			}
			if _, err := l.hub.store.TransferCharacterItemToAccountStash(ctx, member.AccountID, member.CharacterID, ev.ItemInstanceID, ev.StashItemID); err != nil {
				l.hub.metrics.PersistenceErrors.Inc()
				l.log.Error("persist stash item deposit", "account_id", member.AccountID, "character_id", member.CharacterID, "item_instance_id", ev.ItemInstanceID, "stash_item_id", ev.StashItemID, "error", err)
			}
		case "stash_item_withdrawn":
			if ev.ItemInstanceID == "" || ev.StashItemID == "" {
				break
			}
			if _, err := l.hub.store.TransferAccountStashItemToCharacter(ctx, member.AccountID, member.CharacterID, ev.StashItemID, ev.ItemInstanceID); err != nil {
				l.hub.metrics.PersistenceErrors.Inc()
				l.log.Error("persist stash item withdraw", "account_id", member.AccountID, "character_id", member.CharacterID, "item_instance_id", ev.ItemInstanceID, "stash_item_id", ev.StashItemID, "error", err)
			}
		case "stash_gold_deposited":
			if ev.Amount == nil {
				break
			}
			if _, _, err := l.hub.store.TransferCharacterGoldToAccountStash(ctx, member.AccountID, member.CharacterID, *ev.Amount); err != nil {
				l.hub.metrics.PersistenceErrors.Inc()
				l.log.Error("persist stash gold deposit", "account_id", member.AccountID, "character_id", member.CharacterID, "amount", *ev.Amount, "error", err)
			}
		case "stash_gold_withdrawn":
			if ev.Amount == nil {
				break
			}
			if _, _, err := l.hub.store.TransferAccountStashGoldToCharacter(ctx, member.AccountID, member.CharacterID, *ev.Amount); err != nil {
				l.hub.metrics.PersistenceErrors.Inc()
				l.log.Error("persist stash gold withdraw", "account_id", member.AccountID, "character_id", member.CharacterID, "amount", *ev.Amount, "error", err)
			}
		}
		eventSequence++
	}

	hotbarAssignedItems := map[string]struct{}{}
	for _, c := range res.Changes {
		if c.Op == game.OpHotbarUpdate && c.ItemInstanceID != nil {
			hotbarAssignedItems[*c.ItemInstanceID] = struct{}{}
		}
	}
	for _, c := range res.Changes {
		if c.StashTransferID != "" {
			continue
		}
		changeMember := member
		if c.OwnerPlayerID != 0 {
			if ownerMember, ok := membersByPlayerID[c.OwnerPlayerID]; ok {
				changeMember = ownerMember
			}
		}
		switch c.Op {
		case game.OpInventoryAdd:
			if c.Item == nil {
				continue
			}
			location := store.ItemLocationInventory
			if c.Item.Equipped {
				location = store.ItemLocationEquipped
			}
			rolledStats := json.RawMessage(`{}`)
			if payload := c.Item.RollPayload(); payload != nil {
				if raw, err := json.Marshal(payload); err == nil {
					rolledStats = raw
				} else {
					l.hub.metrics.PersistenceErrors.Inc()
					l.log.Error("marshal rolled item payload", "error", err)
				}
			}
			if err := l.hub.store.AddCharacterItem(ctx, store.CharacterItemInstance{
				ID:          c.Item.ItemInstanceID,
				AccountID:   changeMember.AccountID,
				CharacterID: changeMember.CharacterID,
				ItemDefID:   c.Item.ItemDefID,
				Location:    location,
				Slot:        c.Item.Slot,
				Equipped:    c.Item.Equipped,
				RolledStats: rolledStats,
			}); err != nil {
				l.hub.metrics.PersistenceErrors.Inc()
				l.log.Error("persist inventory add", "error", err)
			}
		case game.OpInventoryUpdate:
			if c.Item == nil {
				continue
			}
			if err := l.hub.store.SetCharacterItemEquipped(ctx, changeMember.AccountID, changeMember.CharacterID, c.Item.ItemInstanceID, c.Item.Slot, c.Item.Equipped); err != nil {
				l.hub.metrics.PersistenceErrors.Inc()
				l.log.Error("persist inventory update", "error", err)
			}
		case game.OpInventoryRemove:
			if c.ItemInstanceID == nil {
				continue
			}
			if _, assignedToHotbar := hotbarAssignedItems[*c.ItemInstanceID]; assignedToHotbar {
				continue
			}
			if err := l.hub.store.RemoveCharacterItem(ctx, changeMember.AccountID, changeMember.CharacterID, *c.ItemInstanceID); err != nil {
				l.hub.metrics.PersistenceErrors.Inc()
				l.log.Error("persist inventory remove", "error", err)
			}
		case game.OpEquippedUpdate:
			if c.ItemInstanceID == nil || c.Slot == "" {
				continue
			}
			if err := l.hub.store.SetCharacterItemEquipped(ctx, changeMember.AccountID, changeMember.CharacterID, *c.ItemInstanceID, c.Slot, true); err != nil {
				l.hub.metrics.PersistenceErrors.Inc()
				l.log.Error("persist equipped update", "error", err)
			}
		case game.OpHotbarUpdate:
			if err := l.hub.store.SetCharacterHotbarSlot(ctx, changeMember.AccountID, changeMember.CharacterID, c.SlotIndex, c.ItemInstanceID); err != nil {
				l.hub.metrics.PersistenceErrors.Inc()
				l.log.Error("persist hotbar update", "error", err)
			}
		case game.OpSkillBindingsUpdate:
			if c.SkillBindings == nil {
				continue
			}
			if err := l.hub.store.SetCharacterSkillBindings(ctx, store.CharacterSkillBindings{
				AccountID:         changeMember.AccountID,
				CharacterID:       changeMember.CharacterID,
				FunctionKeys:      c.SkillBindings.FunctionKeys,
				RightClickSkillID: c.SkillBindings.RightClickSkillID,
			}); err != nil {
				l.hub.metrics.PersistenceErrors.Inc()
				l.log.Error("persist skill bindings update", "error", err)
			}
		case game.OpGoldUpdate:
			if c.Gold == nil {
				continue
			}
			if err := l.hub.store.SetCharacterGold(ctx, changeMember.AccountID, changeMember.CharacterID, *c.Gold); err != nil {
				l.hub.metrics.PersistenceErrors.Inc()
				l.log.Error("persist character gold", "error", err)
			}
		case game.OpTeleporterDiscoveryUpdate:
			if c.Discovered {
				if _, err := l.hub.store.AddCharacterWaypoint(ctx, changeMember.CharacterID, c.Level); err != nil {
					l.hub.metrics.PersistenceErrors.Inc()
					l.log.Error("persist character waypoint", "error", err)
				}
			}
		case game.OpShopStockReplace:
			if err := l.hub.store.ReplaceCharacterShopStock(ctx, changeMember.AccountID, changeMember.CharacterID, c.ShopID, c.RefreshKey, storeShopStock(changeMember.AccountID, changeMember.CharacterID, c.ShopStock)); err != nil {
				l.hub.metrics.PersistenceErrors.Inc()
				l.log.Error("persist shop stock replace", "shop_id", c.ShopID, "error", err)
			}
		case game.OpShopStockAvailability:
			if err := l.hub.store.SetCharacterShopStockAvailable(ctx, changeMember.AccountID, changeMember.CharacterID, c.ShopID, c.OfferID, c.Available); err != nil {
				l.hub.metrics.PersistenceErrors.Inc()
				l.log.Error("persist shop stock availability", "shop_id", c.ShopID, "offer_id", c.OfferID, "error", err)
			}
		case game.OpCharacterProgressionUpdate:
			if c.Progression == nil {
				continue
			}
			if err := l.hub.store.UpsertCharacterProgression(ctx, changeMember.AccountID, storeProgressionFromView(changeMember.AccountID, changeMember.CharacterID, *c.Progression)); err != nil {
				l.hub.metrics.PersistenceErrors.Inc()
				l.log.Error("persist character progression", "error", err)
			}
		}
	}

	if err := l.hub.store.TouchSession(ctx, l.sess.ID); err != nil {
		l.hub.metrics.PersistenceErrors.Inc()
	}
	return eventSequence
}

func (l *sessionLoop) currentTick() uint64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.sim.CurrentTick()
}

func (l *sessionLoop) errorEnvelope(code, message string) outEnvelope {
	return outEnvelope{Type: typeError, MessageID: ids.New("msg"), SessionID: l.sess.ID, Tick: l.currentTick(), Payload: errorPayload{Code: code, Message: message}}
}

func (l *sessionLoop) acceptedEnvelope(messageID string, tick uint64, corr string) outEnvelope {
	return outEnvelope{Type: typeIntentAccepted, MessageID: ids.New("msg"), SessionID: l.sess.ID, Tick: tick, CorrelationID: corr, Payload: intentAcceptedPayload{AcceptedMessageID: messageID, ServerTick: tick}}
}

func (l *sessionLoop) rejectedEnvelope(messageID, reason, corr string) outEnvelope {
	l.hub.metrics.RejectedIntents.Inc()
	return outEnvelope{Type: typeIntentRejected, MessageID: ids.New("msg"), SessionID: l.sess.ID, Tick: l.currentTick(), CorrelationID: corr, Payload: intentRejectedPayload{RejectedMessageID: messageID, Reason: reason}}
}

func (l *sessionLoop) clientsForLevelLocked(level int) []*loopClient {
	out := []*loopClient{}
	for _, client := range l.clients {
		if clientLevel, ok := l.sim.PlayerCurrentLevel(client.playerID); ok && clientLevel == level {
			out = append(out, client)
		}
	}
	return out
}

func memberKey(member store.SessionMember) string {
	return member.AccountID + "/" + member.CharacterID
}

func idStr(id uint64) string {
	return fmt.Sprintf("%d", id)
}

func displayNameForMember(member store.SessionMember) string {
	if member.Role == store.SessionMemberHost {
		return "Hero"
	}
	if member.CharacterID == "" {
		return "Guest"
	}
	suffix := member.CharacterID
	if len(suffix) > 6 {
		suffix = suffix[len(suffix)-6:]
	}
	return "Guest " + suffix
}

func killedEventMember(ev game.Event, membersByPlayerID map[uint64]store.SessionMember) (store.SessionMember, bool) {
	entityID := ev.TargetEntityID
	if entityID == "" {
		entityID = ev.EntityID
	}
	playerID, ok := game.ParseEntityID(entityID)
	if !ok {
		return store.SessionMember{}, false
	}
	member, ok := membersByPlayerID[playerID]
	return member, ok
}

func eventEntityPlayerID(ev game.Event) (uint64, bool) {
	if ev.EntityID == "" {
		return 0, false
	}
	return game.ParseEntityID(ev.EntityID)
}

func isCoopSession(sess store.Session) bool {
	return sess.Mode == store.SessionModeCoop || sess.JoinCodeHash != ""
}
