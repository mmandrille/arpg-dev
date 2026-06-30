# v387 Spec — Class Weapon Affinities

Status: Ready for implementation
Date: 2026-06-29
Codename: class-weapon-affinities

## Purpose

Add rollable **class affinities** on weapon/shield templates. Each rolled item carries affinity
rows (e.g. `+10% attack speed (Rogue)`). Server exposes `class_affinity_status` with `active` per
viewing character; client shows green when active, red when inactive. Active affinities affect
authoritative derived/combat stats.

Exemplar families: dagger/rogue atk speed, war hammer/barbarian damage, heraldic shield/non-paladin
atk speed penalty, compound bow/ranger reach, staff/sorcerer max mana.

## Non-goals

- Hard class equip locks; stat requirements only
- New production art (borrow placeholder families)
- Mystery seller / market changes beyond status on listings

## Acceptance criteria

- [ ] `class_affinities` on templates with roll ranges; persisted in `ItemRollPayload`
- [ ] `class_affinity_status` on inventory/loot/shop/stash views
- [ ] Active affinities apply; inactive do not; heraldic penalty only for non-paladins
- [ ] War hammer `damage_percent` scales basic attack and weapon-scaled skills (post-v386)
- [ ] Green/red tooltip lines in inventory, shop, market
- [ ] Five exemplar templates + loot access for lab proof
- [ ] Extended bot scenario cross-class active/inactive proof
- [ ] Focused Go tests + `make validate-shared` + `make client-unit`
