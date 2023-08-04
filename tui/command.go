package tui

func (t *Tui) RenderCommand(command string, cursorPos int) {
	row := t.renderer.Rows - 1
	formatted := padWithSpaces(":" + command, len(command), t.renderer.Cols)
	cursorPos = cursorPos+1

	t.renderer.SetAttribute(attrStatus)
	t.renderer.SetString(row, 0, formatted[:cursorPos])
	t.renderer.SetAttribute(attrActive)
	t.renderer.SetString(row, cursorPos, formatted[cursorPos:cursorPos+1])
	t.renderer.SetAttribute(attrStatus)
	t.renderer.SetString(row, cursorPos+1, formatted[cursorPos+1:])

	t.out.Flush()
}
