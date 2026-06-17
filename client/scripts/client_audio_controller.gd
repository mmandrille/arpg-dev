extends Node
class_name ClientAudioController

const SAMPLE_RATE := 22050
const CUE_DURATION := 0.16
const BOSS_MUSIC_DURATION := 2.4
const AMBIENCE_DURATION := 2.8
const MIN_VOLUME := 0.0
const MAX_VOLUME := 1.0

var master_volume: float = 0.8
var music_volume: float = 0.7
var sfx_volume: float = 0.8
var last_cue: String = ""
var cue_count: int = 0
var boss_music_active: bool = false
var boss_music_intensity: float = 0.0
var boss_music_layer: String = "none"
var last_boss_pattern_id: String = ""
var last_boss_phase_kind: String = ""
var ambient_zone: String = "none"
var ambient_active: bool = false
var last_skill_id: String = ""

var _sfx_player: AudioStreamPlayer
var _music_player: AudioStreamPlayer
var _cue_streams: Dictionary = {}
var _music_stream: AudioStreamWAV
var _ambient_streams: Dictionary = {}


func _ready() -> void:
	_ensure_players()


static func clamp_volume(value: float) -> float:
	return clampf(value, MIN_VOLUME, MAX_VOLUME)


func apply_volumes(master: float, music: float, sfx: float) -> void:
	master_volume = clamp_volume(master)
	music_volume = clamp_volume(music)
	sfx_volume = clamp_volume(sfx)
	_ensure_players()
	_sfx_player.volume_db = _volume_to_db(master_volume * sfx_volume)
	_music_player.volume_db = _volume_to_db(master_volume * music_volume)


func play_cue(cue: String) -> void:
	var normalized := _normalize_cue(cue)
	last_cue = normalized
	cue_count += 1
	if master_volume <= 0.0 or sfx_volume <= 0.0:
		return
	_ensure_players()
	_sfx_player.stream = _stream_for_cue(normalized)
	_sfx_player.play()


func play_movement() -> void:
	play_cue("movement")


func play_attack() -> void:
	play_cue("attack")


func play_skill(skill_id: String = "") -> void:
	last_skill_id = skill_id.strip_edges().to_lower()
	play_cue(_skill_cue(last_skill_id))


func play_heal() -> void:
	play_cue("heal")


func play_damage(local_player: bool) -> void:
	play_cue("player_damage" if local_player else "monster_damage")


func play_kill(is_boss: bool = false) -> void:
	play_cue("boss_kill" if is_boss else "monster_kill")


func play_boss_phase(pattern_id: String = "", phase_kind: String = "") -> void:
	var normalized_pattern := _normalize_boss_value(pattern_id)
	var normalized_phase := _normalize_boss_value(phase_kind)
	last_boss_pattern_id = normalized_pattern
	last_boss_phase_kind = normalized_phase
	boss_music_layer = _boss_music_layer(normalized_pattern, normalized_phase)
	boss_music_intensity = _boss_music_intensity(boss_music_layer)
	play_cue(_boss_phase_cue(normalized_pattern, normalized_phase))
	start_boss_music()


func start_boss_music() -> void:
	boss_music_active = true
	ambient_active = false
	if master_volume <= 0.0 or music_volume <= 0.0:
		return
	_ensure_players()
	_music_player.stream = _boss_music_stream()
	_music_player.play()


func stop_boss_music() -> void:
	boss_music_active = false
	boss_music_intensity = 0.0
	boss_music_layer = "none"
	if _music_player != null:
		_music_player.stop()
	_resume_ambience()


func set_ambient_level(level: int) -> void:
	var next_zone := "town" if level == 0 else "dungeon"
	if ambient_zone == next_zone and (ambient_active or boss_music_active):
		return
	ambient_zone = next_zone
	_resume_ambience()


func get_debug_state() -> Dictionary:
	return {
		"last_cue": last_cue,
		"cue_count": cue_count,
		"boss_music_active": boss_music_active,
		"boss_music_intensity": boss_music_intensity,
		"boss_music_layer": boss_music_layer,
		"last_boss_pattern_id": last_boss_pattern_id,
		"last_boss_phase_kind": last_boss_phase_kind,
		"ambient_zone": ambient_zone,
		"ambient_active": ambient_active,
		"last_skill_id": last_skill_id,
	}


func _ensure_players() -> void:
	if _sfx_player == null:
		_sfx_player = AudioStreamPlayer.new()
		_sfx_player.name = "SfxPlayer"
		add_child(_sfx_player)
	if _music_player == null:
		_music_player = AudioStreamPlayer.new()
		_music_player.name = "MusicPlayer"
		add_child(_music_player)
	_music_player.volume_db = _volume_to_db(master_volume * music_volume)
	_sfx_player.volume_db = _volume_to_db(master_volume * sfx_volume)


func _stream_for_cue(cue: String) -> AudioStreamWAV:
	if not _cue_streams.has(cue):
		_cue_streams[cue] = _make_tone_stream(_cue_frequency(cue), CUE_DURATION, false, _cue_wave(cue))
	return _cue_streams[cue]


func _boss_music_stream() -> AudioStreamWAV:
	if _music_stream == null:
		_music_stream = _make_boss_music_stream()
	return _music_stream


func _resume_ambience() -> void:
	if boss_music_active or ambient_zone == "none":
		ambient_active = false
		return
	ambient_active = true
	if master_volume <= 0.0 or music_volume <= 0.0:
		return
	_ensure_players()
	_music_player.stream = _ambient_stream(ambient_zone)
	_music_player.play()


func _ambient_stream(zone: String) -> AudioStreamWAV:
	if not _ambient_streams.has(zone):
		_ambient_streams[zone] = _make_ambient_stream(zone)
	return _ambient_streams[zone]


func _make_boss_music_stream() -> AudioStreamWAV:
	var stream := _make_tone_stream(110.0, BOSS_MUSIC_DURATION, true, "boss")
	stream.loop_mode = AudioStreamWAV.LOOP_FORWARD
	stream.loop_begin = 0
	stream.loop_end = int(BOSS_MUSIC_DURATION * SAMPLE_RATE)
	return stream


func _make_ambient_stream(zone: String) -> AudioStreamWAV:
	var frequency := 72.0 if zone == "town" else 58.0
	var wave := "town_ambience" if zone == "town" else "dungeon_ambience"
	var stream := _make_tone_stream(frequency, AMBIENCE_DURATION, true, wave)
	stream.loop_mode = AudioStreamWAV.LOOP_FORWARD
	stream.loop_begin = 0
	stream.loop_end = int(AMBIENCE_DURATION * SAMPLE_RATE)
	return stream


func _make_tone_stream(frequency: float, duration: float, looped: bool, wave: String) -> AudioStreamWAV:
	var frame_count: int = max(1, int(duration * SAMPLE_RATE))
	var bytes := PackedByteArray()
	bytes.resize(frame_count * 2)
	for i in frame_count:
		var t := float(i) / float(SAMPLE_RATE)
		var sample: float = _sample_for_wave(wave, frequency, t)
		var envelope: float = _envelope(float(i) / float(frame_count), looped)
		var value: int = int(clampf(sample * envelope, -1.0, 1.0) * 28000.0)
		bytes[i * 2] = value & 0xff
		bytes[i * 2 + 1] = (value >> 8) & 0xff
	var stream := AudioStreamWAV.new()
	stream.format = AudioStreamWAV.FORMAT_16_BITS
	stream.mix_rate = SAMPLE_RATE
	stream.stereo = false
	stream.data = bytes
	return stream


func _sample_for_wave(wave: String, frequency: float, t: float) -> float:
	if wave == "boss":
		return (
			sin(TAU * 55.0 * t) * 0.42
			+ sin(TAU * 82.41 * t) * 0.24
			+ sin(TAU * 110.0 * t) * 0.18
			+ sin(TAU * 146.83 * t) * 0.10
		)
	if wave == "town_ambience":
		return (
			sin(TAU * frequency * t) * 0.32
			+ sin(TAU * (frequency * 1.5) * t) * 0.14
			+ sin(TAU * 0.35 * t) * 0.10
		)
	if wave == "dungeon_ambience":
		return (
			sin(TAU * frequency * t) * 0.28
			+ sin(TAU * (frequency * 0.5) * t) * 0.22
			+ sin(TAU * (frequency * 2.01) * t) * 0.08
		)
	if wave == "noise":
		return sin(TAU * frequency * t) * 0.45 + sin(TAU * frequency * 1.51 * t) * 0.35
	if wave == "down":
		var slide := frequency - (frequency * 0.45 * minf(1.0, t / CUE_DURATION))
		return sin(TAU * slide * t)
	if wave == "up":
		var lift := frequency + (frequency * 0.55 * minf(1.0, t / CUE_DURATION))
		return sin(TAU * lift * t)
	return sin(TAU * frequency * t)


func _envelope(progress: float, looped: bool) -> float:
	if looped:
		return 0.65
	if progress < 0.12:
		return progress / 0.12
	return maxf(0.0, 1.0 - progress)


func _cue_frequency(cue: String) -> float:
	match cue:
		"movement":
			return 180.0
		"attack":
			return 310.0
		"skill":
			return 660.0
		"skill_projectile":
			return 700.0
		"skill_buff":
			return 470.0
		"skill_protection":
			return 560.0
		"skill_revive":
			return 820.0
		"movement_skill":
			return 520.0
		"heal":
			return 740.0
		"player_damage":
			return 150.0
		"monster_damage":
			return 420.0
		"monster_kill":
			return 240.0
		"boss_kill":
			return 90.0
		"boss_phase":
			return 115.0
		"boss_telegraph":
			return 128.0
		"boss_active":
			return 92.0
		"boss_recovery":
			return 180.0
		"boss_summon":
			return 196.0
		"boss_ranged":
			return 260.0
		_:
			return 440.0


func _cue_wave(cue: String) -> String:
	match cue:
		"movement", "monster_damage":
			return "noise"
		"skill", "skill_projectile", "skill_buff", "skill_protection", "skill_revive", "movement_skill", "heal":
			return "up"
		"player_damage", "monster_kill", "boss_kill", "boss_phase", "boss_active":
			return "down"
		"boss_telegraph", "boss_recovery", "boss_summon", "boss_ranged":
			return "up"
		_:
			return "sine"


func _skill_cue(skill_id: String) -> String:
	match skill_id:
		"heal":
			return "heal"
		"leap", "charge", "earthbreaker", "disengage", "teleport":
			return "movement_skill"
		"magic_bolt", "arcane_barrage", "poison_stab", "pinning_shot", "volley":
			return "skill_projectile"
		"rage", "duelist_mark":
			return "skill_buff"
		"holy_shield", "sanctuary":
			return "skill_protection"
		"revive":
			return "skill_revive"
		_:
			return "skill"


func _boss_phase_cue(pattern_id: String, phase_kind: String) -> String:
	if pattern_id.contains("summon"):
		return "boss_summon"
	if pattern_id.contains("lance") or pattern_id.contains("line") or pattern_id.contains("fan"):
		return "boss_ranged"
	match phase_kind:
		"telegraph":
			return "boss_telegraph"
		"active":
			return "boss_active"
		"recovery":
			return "boss_recovery"
		_:
			return "boss_phase"


func _boss_music_layer(pattern_id: String, phase_kind: String) -> String:
	var cue := _boss_phase_cue(pattern_id, phase_kind)
	match cue:
		"boss_summon":
			return "summon"
		"boss_ranged":
			return "ranged"
		"boss_telegraph":
			return "windup"
		"boss_active":
			return "danger"
		"boss_recovery":
			return "release"
		_:
			return "steady"


func _boss_music_intensity(layer: String) -> float:
	match layer:
		"summon":
			return 0.86
		"ranged":
			return 0.78
		"windup":
			return 0.64
		"danger":
			return 1.0
		"release":
			return 0.34
		_:
			return 0.5


func _normalize_boss_value(value: String) -> String:
	return value.strip_edges().to_lower()


func _normalize_cue(cue: String) -> String:
	var normalized := cue.strip_edges().to_lower()
	if normalized == "":
		return "default"
	return normalized


func _volume_to_db(linear: float) -> float:
	if linear <= 0.001:
		return -80.0
	return linear_to_db(linear)
