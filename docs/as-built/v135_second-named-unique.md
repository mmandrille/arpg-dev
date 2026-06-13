# v135 As-Built: Second Named Unique

Date: 2026-06-13
Status: Complete

## What Changed

- Added `stormstring_bow` to `shared/rules/unique_items.v0.json` as a second enabled ready named
  unique.
- The new named unique uses the existing `cave_bow` template and live `stormbound_echo` effect.
- Go tests now assert fixed payload construction for both `embercall_blade` and `stormstring_bow`.
- The deterministic purple town unique chest test now proves both named unique rows are present.
- Python unique-item validator fixtures include a representative second named unique.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestNamedUnique|TestUniqueTestChest'`
- `.venv/bin/python -m pytest tools/test_validate_unique_items.py -q`
- `make maintainability`
- `make ci`

## Notes

- Natural unique drop odds are unchanged. This slice only expands the enabled named unique catalog
  used by the deterministic purple town unique chest path.
