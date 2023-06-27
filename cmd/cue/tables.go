package main

import "text/tabwriter"

type TableOutput struct{ w *tabwriter.Writer }

func (out *TableOutput) basicFooter() {
	out.w.Flush()
}
