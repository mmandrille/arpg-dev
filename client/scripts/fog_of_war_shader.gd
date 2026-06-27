## Canvas fog compositor shader source for FogOfWarOverlay.
class_name FogOfWarShader
extends RefCounted


static func code() -> String:
	return """
shader_type canvas_item;
render_mode blend_mix, unshaded;

uniform float perspective_mode = 0.0;
uniform vec2 center_px = vec2(0.0, 0.0);
uniform vec2 viewport_px = vec2(1.0, 1.0);
uniform vec2 hero_world_xz = vec2(0.0, 0.0);
uniform vec3 camera_origin = vec3(0.0, 0.0, 0.0);
uniform mat4 inv_projection;
uniform mat4 inv_view;
uniform float light_radius = 0.0;
uniform float light_radius_px = 0.0;
uniform float darkness_feather_px = 0.0;
uniform float edge_feather_world = 1.15;
uniform float falloff_power = 2.0;
uniform float darkness_alpha = 1.0;
uniform float organic_edge_px = 0.0;
uniform float organic_edge_amp_world = 0.0;
uniform float organic_edge_segments = 18.0;
uniform float organic_edge_seed = 41.0;
uniform float organic_edge_rotation = 0.0;
uniform int torch_count = 0;
uniform vec2 torch_world_xz[8];
uniform float torch_light_radius = 0.0;
uniform vec4 darkness_color : source_color = vec4(0.0, 0.0, 0.0, 1.0);
uniform float sample_height_count = 1.0;
uniform float sample_height_1 = 1.25;
uniform float sample_height_2 = 2.5;
uniform float sample_height_3 = 3.75;
uniform float height_sample_max_ground_scale = 2.0;

const float ORGANIC_PI = 3.14159265359;
const float ORGANIC_TAU = 6.28318530718;
const float FAR_WORLD_DIST = 100000.0;

float hash1(float n) {
	return fract(sin(n) * 43758.5453123);
}

float smooth_noise(float x) {
	float i = floor(x);
	float f = fract(x);
	float u = f * f * (3.0 - 2.0 * f);
	return mix(hash1(i), hash1(i + 1.0), u);
}

float organic_edge_screen(vec2 delta_px) {
	if (organic_edge_px <= 0.0 || length(delta_px) <= 0.001) {
		return 0.0;
	}
	float angle = fract((atan(delta_px.y, delta_px.x) + ORGANIC_PI) / ORGANIC_TAU + organic_edge_rotation);
	float low = smooth_noise(angle * organic_edge_segments + organic_edge_seed);
	float high = smooth_noise(angle * organic_edge_segments * 2.17 + organic_edge_seed * 3.31);
	float combined = low * 0.72 + high * 0.28;
	return (combined - 0.5) * 2.0 * organic_edge_px;
}

float organic_edge_world(vec2 delta_xz) {
	if (organic_edge_amp_world <= 0.0 || length(delta_xz) <= 0.001) {
		return 0.0;
	}
	float angle = fract((atan(delta_xz.y, delta_xz.x) + ORGANIC_PI) / ORGANIC_TAU + organic_edge_rotation);
	float low = smooth_noise(angle * organic_edge_segments + organic_edge_seed);
	float high = smooth_noise(angle * organic_edge_segments * 2.17 + organic_edge_seed * 3.31);
	float combined = low * 0.72 + high * 0.28;
	return (combined - 0.5) * 2.0 * organic_edge_amp_world;
}

vec3 world_ray_direction(vec2 screen_px) {
	vec2 uv = screen_px / viewport_px;
	uv.y = 1.0 - uv.y;
	vec2 ndc = uv * 2.0 - 1.0;
	vec4 view = inv_projection * vec4(ndc, 1.0, 1.0);
	view.xyz /= view.w;
	vec3 dir_view = normalize(view.xyz);
	return normalize((inv_view * vec4(dir_view, 0.0)).xyz);
}

vec2 plane_xz_at_height(vec2 screen_px, float plane_y) {
	vec3 dir = world_ray_direction(screen_px);
	if (abs(dir.y) <= 0.0001) {
		return hero_world_xz + vec2(FAR_WORLD_DIST, 0.0);
	}
	float t = (plane_y - camera_origin.y) / dir.y;
	if (t < 0.0) {
		return hero_world_xz + vec2(FAR_WORLD_DIST, 0.0);
	}
	vec3 hit = camera_origin + dir * t;
	return hit.xz;
}

vec2 ground_xz_at_screen(vec2 screen_px) {
	return plane_xz_at_height(screen_px, 0.0);
}

float visibility_from_distance(float dist, float effective_radius, float feather) {
	float normalized = dist / max(0.001, effective_radius);
	float visibility = clamp(1.0 - pow(normalized, falloff_power), 0.0, 1.0);
	if (dist > effective_radius) {
		float feather_t = smoothstep(effective_radius, effective_radius + max(0.001, feather), dist);
		visibility = mix(visibility, 0.0, feather_t);
	}
	return visibility;
}

float torch_visibility_at_world_xz(vec2 world_xz) {
	if (torch_count <= 0 || torch_light_radius <= 0.0) {
		return 0.0;
	}
	float best = 0.0;
	for (int i = 0; i < 8; i++) {
		if (i >= torch_count) {
			break;
		}
		vec2 delta = world_xz - torch_world_xz[i];
		float dist = length(delta);
		best = max(best, visibility_from_distance(dist, torch_light_radius, edge_feather_world * 0.55));
	}
	return best;
}

float visibility_at_world_xz(vec2 world_xz) {
	vec2 delta_xz = world_xz - hero_world_xz;
	float dist = length(delta_xz);
	float edge = organic_edge_world(delta_xz);
	float effective_radius = max(0.001, light_radius + edge * 0.45);
	float hero_vis = visibility_from_distance(dist, effective_radius, edge_feather_world);
	float torch_vis = torch_visibility_at_world_xz(world_xz);
	return max(hero_vis, torch_vis);
}

float visibility_isometric(vec2 screen_px) {
	vec2 delta_px = screen_px - center_px;
	float d = length(delta_px);
	float edge = organic_edge_screen(delta_px);
	float effective_radius_px = max(1.0, light_radius_px + edge * 0.45);
	float hero_vis = visibility_from_distance(d, effective_radius_px, darkness_feather_px);
	vec2 world_xz = ground_xz_at_screen(screen_px);
	float torch_vis = torch_visibility_at_world_xz(world_xz);
	return max(hero_vis, torch_vis);
}

float visibility_perspective(vec2 screen_px) {
	return 1.0;
}

void fragment() {
	vec2 screen_px = SCREEN_UV * viewport_px;
	float visibility = perspective_mode > 0.5
		? visibility_perspective(screen_px)
		: visibility_isometric(screen_px);
	float alpha = (1.0 - visibility) * darkness_alpha;
	COLOR = vec4(darkness_color.rgb, alpha);
}
"""
