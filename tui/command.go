package tui

func (t *Tui) RenderCommand(command string) {
	row := t.renderer.Rows - 1
	formatted := padWithSpaces(":" + command, len(command), t.renderer.Cols)

	t.renderer.SetAttribute(attrStatus)
	t.renderer.SetString(row, 0, formatted)
	t.renderer.SetAttribute(attrActive)
	t.renderer.Set(row, len(command)+1, ' ')

	t.out.Flush()
}
