shader_type spatial;
render_mode cull_back, diffuse_burley;

uniform sampler2D ground_texture : source_color, filter_linear_mipmap, repeat_enable;
uniform vec2 ground_uv_scale = vec2(28.0, 18.0);
uniform vec3 soil_color : source_color = vec3(0.45, 0.43, 0.30);
uniform float roughness_value = 0.92;
uniform float blend_width_m = 4.0;
uniform float noise_scale_per_m = 0.45;
uniform float noise_amplitude_m = 0.45;
uniform vec2 plaza_center;
uniform float plaza_radius;
uniform int path_count = 0;
uniform vec4 paths[8];
uniform float path_half_widths[8];

varying vec2 world_xz;

void vertex() {
	world_xz = (MODEL_MATRIX * vec4(VERTEX, 1.0)).xz;
}

float capsule_distance(vec2 p, vec2 a, vec2 b, float half_width) {
	vec2 seg = b - a;
	float t = clamp(dot(p - a, seg) / max(dot(seg, seg), 0.0001), 0.0, 1.0);
	return length(p - (a + seg * t)) - half_width;
}

void fragment() {
	float distance_m = length(world_xz - plaza_center) - plaza_radius;
	for (int i = 0; i < 8; i++) {
		if (i >= path_count) {
			break;
		}
		vec4 path = paths[i];
		distance_m = min(distance_m, capsule_distance(world_xz, path.xy, path.zw, path_half_widths[i]));
	}
	float wave = sin(world_xz.x * noise_scale_per_m) * sin(world_xz.y * noise_scale_per_m * 0.73);
	float edge = max(distance_m + wave * noise_amplitude_m, 0.0);
	float soil = 1.0 - smoothstep(0.0, blend_width_m, edge);
	vec3 grass = texture(ground_texture, UV * ground_uv_scale).rgb;
	ALBEDO = mix(grass, soil_color, soil);
	ROUGHNESS = roughness_value;
}
