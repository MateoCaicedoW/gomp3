package converter

import (
	"gomp3/internal/system/assets"
	"gomp3/internal/system/helpers"

	lucide "github.com/eduardolat/gomponents-lucide"
	"github.com/wawandco/gomui"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

type layout struct {
	Title       string
	Description string
	Yield       Node
}

func (props layout) New() Node {
	return HTML5(
		HTML5Props{
			Title:       props.Title,
			Description: props.Description,
			Language:    "en",
			Head: []Node{

				Link(Rel("stylesheet"), Href(assets.Manager.Path("/public/application.css"))),
				Link(Rel("stylesheet"), Href(assets.Manager.Path("/public/basecoatui.css"))),

				Link(Rel("proconnect"), Href("https://fonts.googleapis.com")),
				Link(Rel("preconnect"), Href("https://fonts.gstatic.com"), Attr("crossorigin", "")),
				Link(Rel("stylesheet"), Href("https://fonts.googleapis.com/css2?family=Noto+Sans:wght@400;500;700&amp;family=Space+Grotesk:wght@300;400;500;600;700&amp;display=swap")),
				Raw(helpers.Importmap()),
				Script(Src(assets.Manager.Path("/public/basecoatui.js")), Defer()),
				Script(Src(assets.Manager.Path("/public/htmx.js")), Defer()),
				Script(Src(assets.Manager.Path("/public/htmx-download.js")), Defer()),
			},
			Body: []Node{

				Div(
					Class("min-h-screen flex flex-col bg-base-100 text-base-content"),
					Header(
						Class("flex items-center justify-end p-4"),

						Div(
							Class("flex items-center gap-2"),
							gomui.LinkButtonEl(
								gomui.ButtonOutline,
								gomui.ButtonDefault,
								true,
								"https://github.com/MateoCaicedoW/gomp3",
								lucide.Github(Class("size-5")),
								Target("_blank"),
							),
							gomui.ThemeToggle("theme-toggle", lucide.Sun(Class("size-5"))),
						),
					),

					props.Yield,
				),

				gomui.DarkModeScript(),
			},
		},
	)
}
