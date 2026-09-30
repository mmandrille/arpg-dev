// Additive soft-glow billboard for combat VFX particles (ADR-0018 P5, v492).
// A round radial falloff on a camera-facing quad, tinted by the particle colour (COLOR, from the
// ParticleProcessMaterial colour ramp) and boosted by glow_energy so the render baseline's glow picks
// it up. No texture needed.
shader_type spatial;
render_mode blend_add, unshaded, cull_disabled, depth_draw_never, shadows_disabled;

uniform float glow_energy : hint_range(0.0, 16.0) = 2.0;
uniform float softness : hint_range(0.05, 1.0) = 0.55;

void vertex() {
	// Billboard that keeps the per-particle scale (like BaseMaterial3D billboard mode "particles").
	mat4 billboard = mat4(
		normalize(INV_VIEW_MATRIX[0]),
		normalize(INV_VIEW_MATRIX[1]),
		normalize(INV_VIEW_MATRIX[2]),
		MODEL_MATRIX[3]);
	billboard = billboard * mat4(
		vec4(length(MODEL_MATRIX[0].xyz), 0.0, 0.0, 0.0),
		vec4(0.0, length(MODEL_MATRIX[1].xyz), 0.0, 0.0),
		vec4(0.0, 0.0, length(MODEL_MATRIX[2].xyz), 0.0),
		vec4(0.0, 0.0, 0.0, 1.0));
	MODELVIEW_MATRIX = VIEW_MATRIX * billboard;
}

void fragment() {
	float d = length(UV * 2.0 - 1.0);
	float falloff = 1.0 - smoothstep(1.0 - softness, 1.0, d);
	ALBEDO = COLOR.rgb * glow_energy;
	ALPHA = falloff * COLOR.a;
}
