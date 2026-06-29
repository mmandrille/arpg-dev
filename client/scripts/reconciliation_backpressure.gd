class_name ReconciliationBackpressure
extends RefCounted


static func should_clear_pending_targets(reconciliation_delta: float, threshold: float) -> bool:
	return reconciliation_delta >= threshold
