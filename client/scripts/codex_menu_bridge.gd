class_name CodexMenuBridge
extends RefCounted

const CodexPanelScript := preload("res://scripts/codex_panel.gd")


static func install(menu_layer: CanvasLayer, back_callback: Callable) -> CodexPanel:
	var codex_panel: CodexPanel = CodexPanelScript.new()
	codex_panel.back_requested.connect(back_callback)
	menu_layer.add_child(codex_panel)
	return codex_panel


static func open_from_main(main_menu: MainMenu, codex_panel: CodexPanel) -> void:
	if main_menu != null:
		main_menu.visible = false
	if codex_panel != null:
		codex_panel.show_codex()


static func back_to_main(main_menu: MainMenu, codex_panel: CodexPanel) -> void:
	if codex_panel != null:
		codex_panel.hide_panel()
	if main_menu != null:
		main_menu.show_menu()


static func bot_select_page(codex_panel: CodexPanel, page_id: String) -> void:
	if codex_panel == null:
		return
	if not codex_panel.visible:
		codex_panel.show_codex()
	codex_panel.select_page(page_id)


static func handle_back(main_menu: MainMenu, codex_panel: CodexPanel) -> bool:
	if codex_panel == null or not codex_panel.visible:
		return false
	back_to_main(main_menu, codex_panel)
	return true
