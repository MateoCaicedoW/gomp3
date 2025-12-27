package converter

import (
	lucide "github.com/eduardolat/gomponents-lucide"
	"github.com/wawandco/gomui"
	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

func indexEl() Node {
	return Main(
		Class("flex-grow flex flex-col items-center justify-start pt-12 pb-20 px-4 sm:px-6"),
		Div(
			Class("w-full max-w-3xl flex flex-col items-center gap-12"),
			// Hero Section
			Div(
				Class("text-center space-y-4"),

				gomui.Badge(gomui.BadgeOutline,
					Span(
						Class("relative flex h-2 w-2"),
						Span(
							Class("animate-ping absolute inline-flex h-full w-full rounded-full bg-primary opacity-75"),
						),
						Span(
							Class("relative inline-flex rounded-full h-2 w-2 bg-sky-500"),
						),
					),
					P(Class(" text-primary text-xs font-bold uppercase tracking-wide"), Text("Youtube MP3 Converter")),
				),
				H2(
					Class("text-4xl md:text-5xl lg:text-6xl leading-tight tracking-tight bg-clip-text  font-extrabold"),
					Text("YouTube to MP3"),
				),
				P(
					Class("text-lg md:text-xl max-w-lg mx-auto font-light"),
					Text("Convert videos to high-quality audio in seconds. No registration required."),
				),
			),

			// Input Section
			Form(
				hx.Post("/convert"),
				hx.Indicator("#loading-indicator"),
				hx.DisabledElt("#convert-button"),
				hx.Ext("htmx-download"),
				Class("w-full flex flex-col sm:flex-row gap-3 p-2 rounded-xl border"),
				Div(
					Class("relative w-full"),
					gomui.InputWithClasses(
						"w-full flex-1 h-full bg-transparent border-none focus:ring-0 pl-10 text-base font-medium shadow-none",
						Type("text"),
						Placeholder("Paste YouTube URL here..."),
						AutoFocus(),
						Name("youtube-url"),
					),

					lucide.Link(Class("size-4 absolute top-1/2 -translate-y-1/2 left-3 ")),
				),

				gomui.ButtonWithClasses(
					" shrink-0 px-12 md:h-14 w-full sm:w-auto flex items-center justify-center gap-2",
					gomui.ButtonPrimary,
					gomui.ButtonLg,
					true,
					ID("convert-button"),
					Type("submit"),
					P(
						Class("font-medium text-lg"),
						Text("Convert"),
					),
					lucide.ArrowRight(Class("size-5")),
				),
			),

			gomui.CardWithClasses(
				"w-full loading-indicator",
				ID("loading-indicator"),
				gomui.CardContent(
					Class("flex items-center justify-between"),

					Div(
						Class("flex items-center gap-3"),
						lucide.RefreshCcw(Class("size-5 animate-spin")),
						P(
							Class("text-sm font-bold"),
							Text("Converting Video to MP3..."),
						),
					),
				),
			),
		),

		Script(Raw(
			`document.addEventListener('htmx:afterRequest', (event) => {
				if (event.detail.xhr.status === 200) {
					document.querySelector('form').reset();
				}
			})`,
		)),
	)
}
