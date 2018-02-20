package main

import (
	"image"

	"github.com/mjl-/duit"
)

func middle(msg *duit.Label, actions ...duit.UI) duit.UI {
	return duit.NewMiddle(duit.SpaceXY(10, 10),
		&duit.Grid{
			Columns: 1,
			Padding: duit.NSpaceXY(1, 0, 2),
			Halign:  []duit.Halign{duit.HalignMiddle},
			Kids: duit.NewKids(
				msg,
				&duit.Box{
					Margin: image.Pt(4, 2),
					Kids:   duit.NewKids(actions...),
				},
			),
		},
	)
}
