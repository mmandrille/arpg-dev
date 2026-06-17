# Headless tests for market listing row helper logic.
# Run via: godot --headless --path client --script res://tests/test_market_listing_rows.gd
extends SceneTree

const MarketListingRowsScript := preload("res://scripts/market_listing_rows.gd")

var _pass: int = 0
var _fail: int = 0


func _initialize() -> void:
	_test_listing_expiration_label()
	_test_debug_rows_expose_expiration()
	if _fail == 0:
		print("[gdtest] PASS: test_market_listing_rows (%d assertions)" % _pass)
		quit(0)
	else:
		print("[gdtest] FAIL: test_market_listing_rows (%d failures, %d assertions)" % [_fail, _pass])
		quit(1)


func _test_listing_expiration_label() -> void:
	var listing := {"expires_at": "2026-06-20T12:30:00Z"}
	_check_eq(MarketListingRowsScript.expiration_label(listing, _unix("2026-06-18 12:30:00")), "Expires in 2d 0h", "future duration label")
	_check_eq(MarketListingRowsScript.expiration_label(listing, _unix("2026-06-20 12:29:30")), "Expires in <1m", "sub-minute label")
	_check_eq(MarketListingRowsScript.expiration_label(listing, _unix("2026-06-20 12:31:00")), "Expired", "expired label")
	_check_eq(MarketListingRowsScript.expiration_label({}), "", "missing expires_at hides label")
	_check_true(MarketListingRowsScript.expiration_label({"expires_at": "not-a-date"}).begins_with("Expires not-a-date"), "malformed date fallback")


func _test_debug_rows_expose_expiration() -> void:
	var rows: Array = MarketListingRowsScript.debug_listing_rows([{
		"listing_id": "listing-1",
		"item_def_id": "cave_mail",
		"seller_account_id": "seller-account",
		"price_gold": 25,
		"expires_at": "2026-06-20T12:30:00Z",
	}], func(_rec): return ["Base Armor: +3"])
	_check_eq(rows.size(), 1, "one debug row")
	var row := rows[0] as Dictionary
	_check_eq(str(row.get("expires_at", "")), "2026-06-20T12:30:00Z", "debug expires_at")
	_check_true(bool(row.get("expiration_visible", false)), "debug expiration visible")
	_check_true(str(row.get("expiration_label", "")).begins_with("Expires"), "debug expiration label")
	_check_true((row.get("stat_lines", []) as Array).has("Base Armor: +3"), "debug stat lines preserve callable")


func _unix(value: String) -> float:
	return Time.get_unix_time_from_datetime_string(value)


func _check_true(cond: bool, msg: String) -> void:
	if cond:
		_pass += 1
	else:
		_fail += 1
		push_error("assertion failed: " + msg)


func _check_eq(got, want, msg: String) -> void:
	_check_true(got == want, "%s got=%s want=%s" % [msg, str(got), str(want)])
