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
	Position     gruid.Point
	Cell         gruid.Cell
	Name         string
	Persona      string
	LastResponse string
}

const (
	TextRows       = 15
	PlayAreaWidth  = 187
	PlayAreaHeight = 32
	HelpText       = `
		Autism Village v0.1
		Do not play if you are socially well-adjusted.
		Arrow keys - move. Enter - interact. '[', ']' - Scroll text area left or right.
    `
)

// Model implements gruid.Model interface and represents the
// application's state.
type Model struct {
	grid              gruid.Grid // user interface grid
	interactingEntity *Entity
	convoOptions      string
	pager             *ui.Pager
	convo             ConvoGenerator
	player            Entity
	entities          []Entity
	// other fields with the state of the application
}

func (m *Model) PlayRange() gruid.Range {
	return gruid.Range{
		Min: m.grid.Range().Min.Shift(1, TextRows),
		Max: m.grid.Range().Max.Shift(-1, -1),
	}
}

func (m *Model) interactWithEntity(input string) {
	m.conductConversation(input)
}

func (m *Model) Update(msg gruid.Msg) gruid.Effect {
	// Update your application's state in response to messages.
	switch msg.(type) {
	case gruid.MsgKeyDown:
		msg := msg.(gruid.MsgKeyDown)
		return m.handleKeyDown(msg)
	}
	return nil
}

func (m *Model) handleKeyDown(msg gruid.MsgKeyDown) gruid.Effect {
	key := msg.Key
	speed := 1
	if msg.Mod&gruid.ModShift != 0 {
		speed = 5
	}
	switch key {
	case gruid.KeyArrowLeft:
		m.player.Position.X -= speed
	case gruid.KeyArrowRight:
		m.player.Position.X += speed
	case gruid.KeyArrowDown:
		m.player.Position.Y += speed
	case gruid.KeyArrowUp:
		m.player.Position.Y -= speed
	case gruid.KeyEnter:
		closeToPlayer := gruid.Range{
			Min: m.player.Position.Shift(-1, -1),
			Max: m.player.Position.Shift(2, 2),
		}
		for _, entity := range m.entities {
			if entity.Position.In(closeToPlayer) {
				m.interactingEntity = &entity
				var cmd gruid.Cmd = func() gruid.Msg {
					m.interactWithEntity("")
					return nil
				}
				return cmd
			}
		}
	case "[", "]":
		m.pager.Update(msg)
	case "1", "2", "3":
		return m.handleConversationOption(string(key))
	case "q":
		return gruid.End()
	}
	playRange := m.PlayRange()
	if m.player.Position.X < playRange.Min.X {
		m.player.Position.X = playRange.Min.X
	}
	if m.player.Position.X >= playRange.Max.X {
		m.player.Position.X = playRange.Max.X - 1
	}
	if m.player.Position.Y < playRange.Min.Y {
		m.player.Position.Y = playRange.Min.X
	}
	if m.player.Position.Y >= playRange.Max.Y {
		m.player.Position.Y = playRange.Max.Y - 1
	}
	return nil
}

func (m *Model) handleConversationOption(chosenOption string) gruid.Effect {
	if m.convoOptions == "" {
		m.displayText("ERROR: no ongoing conversation")
		return nil
	}
	searchString := fmt.Sprintf("%v. ", chosenOption)
	startIndex := strings.Index(m.convoOptions, searchString)
	if startIndex == -1 {
		m.displayText(fmt.Sprintf("ERROR: no such conversation response: %v", chosenOption))
		return nil
	}
	startIndex += len(searchString)
	endIndex := strings.Index(m.convoOptions[startIndex:], "\n")
	if endIndex == -1 {
		endIndex = len(m.convoOptions)
	} else {
		endIndex += startIndex
	}
	response := m.convoOptions[startIndex:endIndex]
	var cmd gruid.Cmd = func() gruid.Msg {
		m.interactWithEntity(response)
		return nil
	}
	return cmd
}

func (m *Model) displayText(text string) {
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

func (s styleManager) GetStyle(style gruid.Style) tcell.Style {
	return tcell.StyleDefault.Foreground(tcell.Color(style.Fg))
}

func Run() {
	m := Model{
		convo: NewConvo(),
		grid:  gruid.NewGrid(PlayAreaWidth, PlayAreaHeight+TextRows),
		pager: ui.NewPager(ui.PagerConfig{
			Grid: gruid.NewGrid(PlayAreaWidth, TextRows),
			Keys: ui.PagerKeys{
				Left:  []gruid.Key{"["},
				Right: []gruid.Key{"]"},
			},
			Style: ui.PagerStyle{},
		}),
		player: Entity{
			Position: gruid.Point{X: 8, Y: TextRows + 8},
			Cell: gruid.Cell{Rune: '@', Style: gruid.Style{
				Fg: 0xF,
			}},
			Name: "Player",
		},
		entities: []Entity{
			{
				Position: gruid.Point{X: 24, Y: 24},
				Cell:     gruid.Cell{Rune: 'ǭ'},
				Name:     "Hendry",
				Persona: `
					Your name is Hendry and you are a Shitzu dog who speaks English very poorly.
					You are angry because a human took your stick. You suspect it was me.
					You are also missing your owner, a man named Jello.
				`,
			},
			{
				Position: gruid.Point{X: 32, Y: 32},
				Cell:     gruid.Cell{Rune: 'ʝ'},
				Name:     "Gembo",
				Persona: `
					You are a smooth talking man named Gembo, 40 years of age, and you talk like you're still stuck in the 30s.
					You are originally from the Philippines.
				`,
			},
		},
	}

	m.displayText(HelpText)
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
