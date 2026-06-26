#!/usr/bin/env bash
# Shared Godot CLI flags for headless CI, unit gates, and bot-client automation.
# Interactive play uses project.godot forward_plus; automated paths force
# gl_compatibility so CI does not require forward_plus GPU features.
GODOT_HEADLESS_FLAGS="${GODOT_HEADLESS_FLAGS:---headless --rendering-method gl_compatibility}"
