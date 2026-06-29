# --- CI + Benchmark -----------------------------------------------------------
.PHONY: maintainability check-progress-dashboard ci ci-full benchmark
maintainability: ## Run maintainability ratchet checks
	./scripts/check-file-size-ratchet.sh
	python3 ./scripts/check-extraction-coupling-ratchet.py
	./scripts/check-progress-dashboard.sh

check-progress-dashboard: ## Validate PROGRESS.md dashboard and progress archive links
	./scripts/check-progress-dashboard.sh

ci: ## Run the fast local CI suite (ci scenario pack; quiet; VERBOSE=1 for full logs)
	ARPG_CI_SCENARIO=ci ARPG_ADDR="$(CI_ADDR)" BASE_URL="$(CI_BASE_URL)" ./scripts/ci.sh

ci-full: ## Run the full scenario matrix (all protocol + client bots; ~20+ min)
	ARPG_CI_SCENARIO=all ARPG_ADDR="$(CI_ADDR)" BASE_URL="$(CI_BASE_URL)" ./scripts/ci.sh

benchmark: ## Run perf benchmark scenarios (ci_tier=benchmark) with ARPG_PERF_DEBUG=1 and generate a report
	BENCHMARK_OUT="$(BENCHMARK_OUT)" ARPG_ADDR="$(CI_ADDR)" BASE_URL="$(CI_BASE_URL)" ./scripts/benchmark.sh

perf-analyze: ## Parse a real play-debug log: make perf-analyze LOG=/tmp/arpg-perf.log [OUT=report.txt]
	@if [[ -z "$(LOG)" ]]; then echo "usage: make perf-analyze LOG=/tmp/arpg-perf.log [OUT=path]"; exit 2; fi
	$(PY) -m tools.bot.benchmark_report --play-log "$(LOG)" $(if $(OUT),--out "$(OUT)",)
