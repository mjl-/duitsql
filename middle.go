package main

import (
	"image"

	"github.com/mjl-/duit"
)

func middle(msg *duit.Label, actions ...duit.UI) duit.UI {
	return duit.NewMiddle(
		&duit.Box{
			Padding: duit.SpaceXY(10, 10),
			Kids: duit.NewKids(
				&duit.Grid{
					Columns: 1,
					Padding: duit.NSpace(1, duit.SpaceXY(4, 2)),
					Halign:  []duit.Halign{duit.HalignMiddle},
					Kids: duit.NewKids(
						msg,
						&duit.Box{
							Margin: image.Pt(4, 2),
							Kids:   duit.NewKids(actions...),
						},
					),
				},
			),
		},
	)
}
