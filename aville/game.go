package aville

import (
	"codeberg.org/anaseto/gruid"
	gruidtcell "codeberg.org/anaseto/gruid-tcell"
	"codeberg.org/anaseto/gruid/ui"
	"context"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"log"
	"strings"
)

type Entity struct {
	Position gruid.Point
	Cell     gruid.Cell
	Prompt   string
}

const (
	TextRows       = 15
	PlayAreaWidth  = 187
	PlayAreaHeight = 32
)

// Model implements gruid.Model interface and represents the
// application's state.
type Model struct {
	grid     gruid.Grid // user interface grid
	pager    *ui.Pager
	convo    ConvoGenerator
	player   Entity
	entities []Entity
	// other fields with the state of the application
}

func (m *Model) PlayRange() gruid.Range {
	return gruid.Range{
		Min: m.grid.Range().Min.Shift(1, 1+TextRows),
		Max: m.grid.Range().Max.Shift(-1, -1),
	}
}

func (m *Model) Update(msg gruid.Msg) gruid.Effect {
	// Update your application's state in response to messages.
	switch msg.(type) {
	case gruid.MsgKeyDown:
		switch msg.(gruid.MsgKeyDown).Key {
		case gruid.KeyArrowLeft:
			m.player.Position.X -= 1
			if !m.player.Position.In(m.PlayRange()) {
				m.player.Position.X += 1
			}
		case gruid.KeyArrowRight:
			m.player.Position.X += 1
			if !m.player.Position.In(m.PlayRange()) {
				m.player.Position.X -= 1
			}
		case gruid.KeyArrowDown:
			m.player.Position.Y += 1
			if !m.player.Position.In(m.PlayRange()) {
				m.player.Position.Y -= 1
			}
		case gruid.KeyArrowUp:
			m.player.Position.Y -= 1
			if !m.player.Position.In(m.PlayRange()) {
				m.player.Position.Y += 1
			}
		case gruid.KeyEnter:
			closeToPlayer := gruid.Range{
				Min: m.player.Position.Shift(-1, -1),
				Max: m.player.Position.Shift(2, 2),
			}
			for _, entity := range m.entities {
				if entity.Position.In(closeToPlayer) {
					m.setPagerText(fmt.Sprintf("Generating conversation for %c...", entity.Cell.Rune))
					go func() {
						var response string
						var err error
						if response, err = m.convo.GenerateOptionsAndResponses(entity.Prompt); err != nil {
							m.setPagerText(fmt.Sprintf("Error generating convo: \n%v", err))
							return
						}
						m.setPagerText(response)
					}()
				}
			}
		case "q":
			return gruid.End()
		}
	}
	return nil
}

func (m *Model) setPagerText(text string) {
	var lines []ui.StyledText
	for _, line := range strings.Split(text, "\n") {
		lines = append(lines, ui.Text(strings.TrimSpace(line)))
	}
	m.pager.SetLines(lines)
}

func (m *Model) drawEntity(entity *Entity) {
	m.grid.Set(
		entity.Position,
		entity.Cell,
	)
}

func (m *Model) Draw() gruid.Grid {
	// Write your rendering into the grid and return it.
	m.grid.Fill(gruid.Cell{
		Rune:  ' ',
		Style: gruid.Style{},
	})
	m.grid.Iter(func(point gruid.Point, cell gruid.Cell) {
		if (point.X == 0 || point.X == PlayAreaWidth-1) && point.Y > TextRows {
			m.grid.Set(point, gruid.Cell{
				Rune:  '┃',
				Style: gruid.Style{},
			})
		}
		if point.Y == TextRows || point.Y == PlayAreaHeight+TextRows-1 {
			m.grid.Set(point, gruid.Cell{
				Rune:  '━',
				Style: gruid.Style{},
			})
		}
	})
	m.drawEntity(&m.player)
	for _, entity := range m.entities {
		m.drawEntity(&entity)
	}
	m.grid.Copy(m.pager.Draw())
	return m.grid
}

type styleManager struct{}

func (s styleManager) GetStyle(_ gruid.Style) tcell.Style {
	return tcell.StyleDefault
}

func Run() {
	m := Model{
		convo: NewConvo(),
		grid:  gruid.NewGrid(PlayAreaWidth, PlayAreaHeight+TextRows),
		pager: ui.NewPager(ui.PagerConfig{
			Grid: gruid.NewGrid(PlayAreaWidth, TextRows),
			Keys: ui.PagerKeys{
				PageDown: []gruid.Key{"."},
				PageUp:   []gruid.Key{","},
			},
			Style: ui.PagerStyle{},
		}),
		player: Entity{
			Position: gruid.Point{X: 8, Y: TextRows + 8},
			Cell:     gruid.Cell{Rune: '😜'},
		},
		entities: []Entity{
			{
				Position: gruid.Point{X: 24, Y: 24},
				Cell:     gruid.Cell{Rune: '🐶'},
				Prompt: `Your name is Hendry and you are a Shitzu dog who speaks English very eloquently.
                        You are angry because a human took your stick. You suspect it was me.
                        Generate three very different conversation options I could say to you.
                        Please number them 1-3 and don't make sentences longer than 187 characters.`,
			},
		},
	}
	m.setPagerText(`
    Autism Village v0.1
    Do not play if you are socially well-adjusted.
    Arrow keys - move. Enter - interact.
    `)
	// Specify a driver among the provided ones.
	driver := gruidtcell.NewDriver(gruidtcell.Config{
		StyleManager: styleManager{},
		DisableMouse: false,
		RuneManager:  nil,
		Tty:          nil,
	})
	app := gruid.NewApp(gruid.AppConfig{
		Driver: driver,
		Model:  &m,
	})
	// Start the main loop of the application.
	if err := app.Start(context.Background()); err != nil {
		log.Fatal(err)
	}
}
