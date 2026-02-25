package view

import "github.com/Curt-Park/claudeview/internal/ui"

// RowBuilder builds a table row from a slice of items at a given index.
type RowBuilder[T any] func(items []T, index int, flatMode bool) ui.Row

// ResourceView is a generic scrollable table view for any resource type.
type ResourceView[T any] struct {
	Table    ui.TableView
	FlatMode bool
	Items    []T
	cols     []ui.Column
	flatCols []ui.Column // nil means flat mode uses the same columns
	rowFunc  RowBuilder[T]
}

// NewResourceView creates a ResourceView with the given normal columns, optional flat-mode
// columns, row builder, and initial dimensions.
func NewResourceView[T any](cols, flatCols []ui.Column, rowFunc RowBuilder[T], w, h int) *ResourceView[T] {
	return &ResourceView[T]{
		Table:    ui.NewTableView(cols, w, h),
		cols:     cols,
		flatCols: flatCols,
		rowFunc:  rowFunc,
	}
}

// SetData updates the items and rebuilds the table rows.
func (v *ResourceView[T]) SetData(items []T) {
	v.Items = items
	if v.FlatMode && v.flatCols != nil {
		v.Table.Columns = v.flatCols
	} else {
		v.Table.Columns = v.cols
	}
	rows := make([]ui.Row, len(items))
	for i := range items {
		rows[i] = v.rowFunc(items, i, v.FlatMode)
	}
	v.Table.SetRows(rows)
}

// Sync updates dimensions, nav state, flat mode, and data in one call, then returns the
// updated TableView ready to be assigned to AppModel.Table.
func (v *ResourceView[T]) Sync(items []T, w, h, sel, off int, filter string, flat bool) ui.TableView {
	v.Table.Width = w
	v.Table.Height = h
	v.Table.Selected = sel
	v.Table.Offset = off
	v.Table.Filter = filter
	v.FlatMode = flat
	v.SetData(items)
	return v.Table
}
