package main

import (
	"fmt"
	"math/rand"
	"strings"
	"syscall/js"
	"time"
)

// Cell represents actual value of the cell.
type Cell int

// Who is a flag showing whom to belong items in the game.
type Who int

const (
	CellEmpty Cell = iota
	CellShip
	// Also a check `Cell < CellSmoke` can be used to detect a cell that has not been hit.
	CellSmoke // The empty cells near sunk ships are converted to this one, making it easier to find the next place to hit.
	CellMiss  // Empty cell that was hit.
	CellHit   // Ship cell that was hit.
)

const (
	WhoSelf Who = iota
	WhoThem
)

// TD* constants represent values the cell may display.
const (
	TDEmpty       = "empty"       // nothing
	TDShadow      = "shadow"      // still hidden
	TDUndercursor = "undercursor" // under cursor (only can be added to the shadow cells)
	TDMiss        = "miss"        // uncovered cell -- miss
	TDHit         = "hit"         // uncovered cell -- hit
	TDDebris      = "debris"      // uncovered cell -- sunk ship, i.e. debris
	TDShip        = "ship"
)

const (
	EmojiMiss   = "🌊"
	EmojiHit    = "🔥"
	EmojiDebris = "💩"
)

const (
	GameSize    = 8
	MaxShipSize = 4
)

var allDirections = []xy{{0, -1}, {-1, 0}, {0, 1}, {1, 0}}

type Game struct {
	doc          js.Value
	board        [][][]Cell // Who/Y/X
	shipCount    []int
	enemyPrevHit xy // Previous hit of the enemy or {-1,-1}.

	// HTML elements
	tableThem       js.Value
	setShipButton   js.Value
	startGameButton js.Value

	cellOverListener       *EventListener
	cellOutListener        *EventListener
	cellClickListener      *EventListener
	setShipClickListener   *EventListener
	startGameClickListener *EventListener
}

type xy struct {
	x int
	y int
}

func (p xy) String() string {
	return fmt.Sprintf("(%d,%d)", p.x, p.y)
}

func (p xy) plus(dp xy) xy {
	return xy{p.x + dp.x, p.y + dp.y}
}

type xyter struct {
	p     xy
	p0    xy
	max   xy
	moved bool
}

func iter(p0, max xy) *xyter {
	return &xyter{
		p:     p0,
		p0:    p0,
		max:   max,
		moved: false,
	}
}

func (x *xyter) next() {
	x.moved = true
	x.p.x++
	if x.p.x >= x.max.x {
		x.p.x = 0
		x.p.y++
		if x.p.y >= x.max.y {
			x.p.y = 0
		}
	}
}

func (x *xyter) more() bool {
	return !(x.moved && x.p == x.p0)
}

func (w Who) String() string {
	if w == WhoSelf {
		return "self"
	}
	return "them"
}

func NewGame() (*Game, error) {
	doc := js.Global().Get("document")
	if !doc.Truthy() {
		return nil, fmt.Errorf("cannot get document")
	}
	g := &Game{
		doc:          doc,
		enemyPrevHit: xy{-1, -1},
	}
	g.buildEventListeners()
	if err := g.buildLayout(); err != nil {
		return nil, err
	}
	if err := g.buildGrid(WhoSelf); err != nil {
		return nil, err
	}
	if err := g.buildGrid(WhoThem); err != nil {
		return nil, err
	}
	if err := g.buildControls(); err != nil {
		return nil, err
	}
	if err := g.clear(); err != nil {
		return nil, err
	}
	return g, nil
}

func (g *Game) buildEventListeners() {
	g.cellOverListener = newEventListener("mouseover", func(this, evt js.Value) any {
		target := evt.Get("target")
		clist, n := getClassList(target)
		if n > 0 && clist.Call("contains", TDShadow).Bool() {
			if p0, err := getCellXY(WhoThem, target); err != nil {
				fmt.Println(err.Error())
			} else if g.cell(WhoThem, p0) < CellSmoke {
				clist.Call("add", TDUndercursor)
			}
		}
		return nil
	})
	g.cellOutListener = newEventListener("mouseout", func(this, evt js.Value) any {
		target := evt.Get("target")
		clist, n := getClassList(target)
		if n > 0 {
			clist.Call("remove", TDUndercursor)
		}
		return nil
	})
	g.cellClickListener = newEventListener("click", func(this, evt js.Value) any {
		target := evt.Get("target")
		clist, nelt := getClassList(target)
		if nelt == 0 || !clist.Call("contains", TDShadow).Bool() {
			return nil
		}
		// TODO: do not fire on smoke cell.
		if err := g.actionOnFire(target); err != nil {
			fmt.Println(err.Error())
			return nil
		}
		return nil
	})
	g.setShipClickListener = newEventListener("click", func(this, evt js.Value) any {
		if err := g.placeAllShips(); err != nil {
			g.log("placeAllShips failed: %v -- try again!!!", err)
			g.clear()
			return nil
		}
		g.log("All ships are placed.  Press Start to start the game.")
		return nil
	})
	g.startGameClickListener = newEventListener("click", func(this, evt js.Value) any {
		g.log("Start Game pressed.")
		g.setShipButton.Set("disabled", true)
		g.startGameButton.Set("disabled", true)
		g.cellOutListener.Add(g.tableThem)
		g.cellOverListener.Add(g.tableThem)
		g.cellClickListener.Add(g.tableThem)
		return nil
	})
}

func (g *Game) buildLayout() error {
	board, err := g.getElementByID("board")
	if err != nil {
		return err
	}
	grids := g.createElement("div")
	grids.Set("id", "grids")
	board.Call("append", grids)

	self := g.createElement("span")
	self.Set("id", "self")
	grids.Call("append", self)

	them := g.createElement("span")
	them.Set("id", "them")
	grids.Call("append", them)

	controls := g.createElement("div")
	controls.Set("id", "controls")
	board.Call("append", controls)

	output := g.createElement("textarea")
	output.Set("id", "output")
	output.Set("cols", 70)
	board.Call("append", output)
	return nil
}

func (g *Game) buildControls() error {
	// Set listeners on buttons.
	controls, err := g.getElementByID("controls")
	if err != nil {
		return err
	}
	g.setShipButton = g.createElement("button")
	g.setShipButton.Call("append", "Set ships")
	controls.Call("append", g.setShipButton)
	g.setShipClickListener.Add(g.setShipButton)

	g.startGameButton = g.createElement("button")
	g.startGameButton.Call("append", "Start game")
	controls.Call("append", g.startGameButton)
	g.startGameButton.Set("disabled", true)
	g.startGameClickListener.Add(g.startGameButton)

	testButton := g.createElement("button")
	testButton.Set("id", "test-button")
	testButton.Call("append", "Test")
	controls.Call("append", testButton)
	testButtonCount := 0
	testButtonListener := newEventListener("click", func(this, evt js.Value) any {
		testButtonCount++
		fmt.Printf("testButton #%d start\n", testButtonCount)
		time.Sleep(time.Second * 3)
		fmt.Printf("testButton #%d finish\n", testButtonCount)
		return nil
	})
	testButtonListener.Add(testButton)
	return nil
}

func (g *Game) log(format string, args ...any) {
	ta, err := g.getElementByID("output")
	if err != nil {
		fmt.Println(err)
		return
	}
	s := fmt.Sprintf(format, args...)
	old := ta.Get("innerText").String()
	if len(old) > 0 {
		s = "\n" + s
	}
	ta.Set("innerText", old+s)
}

func (g *Game) clear() error {
	fmt.Println("clear called.")
	g.cellOverListener.Remove(g.tableThem)
	g.cellOutListener.Remove(g.tableThem)
	g.board = make([][][]Cell, 2)
	g.shipCount = make([]int, 2)
	if err := g.clearBoard(WhoSelf); err != nil {
		return err
	}
	if err := g.clearBoard(WhoThem); err != nil {
		return err
	}
	return nil
}

func (g *Game) clearBoard(who Who) error {
	tdClass := []string{TDEmpty, TDShadow}
	g.board[who] = make([][]Cell, GameSize)
	for y := 0; y < GameSize; y++ {
		g.board[who][y] = make([]Cell, GameSize)
		for x := 0; x < GameSize; x++ {
			td, err := g.tdCell(who, xy{x, y})
			if err != nil {
				return err
			}
			td.Set("className", tdClass[who])
		}
	}
	return nil
}

// tdCell returns html TD element corresponding to (x,y) on the self/non-self grid.
func (g *Game) tdCell(who Who, p xy) (js.Value, error) {
	id := fmt.Sprintf("%s%d-%d", who, p.x, p.y)
	return g.getElementByID(id)
}

// cell returns the value of the element corresponding to (x,y) on the self/non-self grid.
// the x, y args can be out of bounds.
func (g *Game) cell(who Who, p xy) Cell {
	if p.x < 0 || p.x >= GameSize || p.y < 0 || p.y >= GameSize {
		return CellMiss
	}
	return g.board[who][p.y][p.x]
}

func (g *Game) setCell(who Who, p xy, value Cell) {
	if p.x < 0 || p.x >= GameSize || p.y < 0 || p.y >= GameSize {
		return
	}
	g.board[who][p.y][p.x] = value
}

func (g *Game) getElementByID(id string) (js.Value, error) {
	elt := g.doc.Call("getElementById", id)
	if !elt.Truthy() {
		return js.Undefined(), fmt.Errorf("cannot find elt with id %q", id)
	}
	return elt, nil
}

func (g *Game) createElement(typ string) js.Value {
	return g.doc.Call("createElement", typ)
}

// Build a grid in the DOM element with id=grid.
func (g *Game) buildGrid(who Who) error {
	grid, err := g.getElementByID(who.String())
	if err != nil {
		return err
	}
	table := g.createElement("table")
	if who == WhoThem {
		g.tableThem = table
	}
	grid.Call("append", table)
	tr := g.createElement("tr")
	table.Call("append", tr)
	tr.Call("append", g.createElement("th"))
	for x := 0; x < GameSize; x++ {
		th := g.createElement("th")
		tr.Call("append", th)
		txt := string([]byte{byte('A') + byte(x)})
		th.Call("append", txt)
	}
	for y := 0; y < GameSize; y++ {
		tr := g.createElement("tr")
		table.Call("append", tr)
		th := g.createElement("th")
		tr.Call("append", th)
		th.Call("append", fmt.Sprint(GameSize-y))
		for x := 0; x < GameSize; x++ {
			td := g.createElement("td")
			td.Set("id", fmt.Sprintf("%s%d-%d", who.String(), x, y))
			tr.Call("append", td)
		}
	}

	tr = g.createElement("tr")
	table.Call("append", tr)
	th := g.createElement("th")
	th.Set("id", fmt.Sprintf("%s-stat", who.String()))
	th.Set("colSpan", GameSize+1)
	tr.Call("append", th)
	th.Call("append", ".")
	return nil
}

// Attempt to set ships.
func (g *Game) placeAllShips() error {
	fmt.Println("placeAllShips called.")
	if err := g.clear(); err != nil {
		return err
	}
	if err := g.placeShips(WhoThem); err != nil {
		return err
	}
	if err := g.placeShips(WhoSelf); err != nil {
		return err
	}
	g.startGameButton.Set("disabled", false)
	return nil
}

func (g *Game) startGame() error {
	g.cellOverListener.Add(g.tableThem)
	g.cellOutListener.Add(g.tableThem)
	return nil
}

func (g *Game) placeShips(who Who) error {
	if err := g.clearBoard(who); err != nil {
		return err
	}
	for shipSize, nShips := MaxShipSize, 1; shipSize > 0; shipSize, nShips = shipSize-1, nShips+1 {
		for i := 0; i < nShips; i++ {
			var dir xy
			dir.x = rand.Intn(2)
			dir.y = 1 - dir.x
			if err := g.placeShip(who, shipSize, dir); err != nil {
				dir.x, dir.y = dir.y, dir.x
				if err := g.placeShip(who, shipSize, dir); err != nil {
					return err
				}
			}
			g.shipCount[who] += shipSize
		}
	}
	// Clear temporary smoke cells around ships with empty.
	for y := 0; y < GameSize; y++ {
		for x := 0; x < GameSize; x++ {
			if g.cell(who, xy{x, y}) == CellSmoke {
				g.setCell(who, xy{x, y}, CellEmpty)
			}
		}
	}
	return g.showStat(who)
}

func (g *Game) placeShip(who Who, shipSize int, dir xy) error {
	// Beyond the max position of the ship.
	maxp := xy{
		GameSize - dir.x*(shipSize-1),
		GameSize - dir.y*(shipSize-1),
	}
	// Initial placement of the ship.
	p0 := xy{rand.Intn(maxp.x), rand.Intn(maxp.y)}
	fmt.Printf("placing the ship %s of size %d, initial place %v, maxp %v\n", who, shipSize, p0, maxp)
	for it := iter(p0, maxp); it.more(); it.next() {
		// Check that all cells are empty at the location of the ship.
		p := it.p
		placeFound := true
		for i, pt := 0, p; i < shipSize; i++ {
			if g.cell(who, pt) != CellEmpty {
				placeFound = false
				break
			}
			pt = pt.plus(dir)
		}
		if !placeFound {
			continue
		}
		// Place the ship here and mark all adjacent cells as busy with CellSmoke.
		// They will be cleared (set to CellEmpty) once all ships are placed.
		g.setCell(who, p.plus(xy{-dir.x, -dir.y}), CellSmoke)
		fmt.Printf("ship %s of size %d is placed at %v dir %v.\n", who, shipSize, p, dir)
		for i := 0; i < shipSize; i++ {
			if who == WhoSelf {
				td, err := g.tdCell(who, p)
				if err != nil {
					return err
				}
				td.Set("className", TDShip)
			}
			g.setCell(who, p, CellShip)
			g.setCell(who, p.plus(xy{dir.y, dir.x}), CellSmoke)
			g.setCell(who, p.plus(xy{-dir.y, -dir.x}), CellSmoke)
			p = p.plus(dir)
		}
		g.setCell(who, p, CellSmoke)
		return nil
	}
	// We could not place the ship.
	return fmt.Errorf("cannot place %s ship of size %d with dir %v -- all cells are busy", who, shipSize, dir)
}

func getCellXY(who Who, where js.Value) (xy, error) {
	var p0 xy
	id := where.Get("id").String()
	s := strings.TrimPrefix(id, who.String())
	n, err := fmt.Sscanf(s, "%d-%d", &p0.x, &p0.y)
	if err != nil || n != 2 {
		return p0, fmt.Errorf("could not read x,y of the cell id=%s: n=%d, err=%v", id, n, err)
	}
	return p0, nil
}

func (g *Game) actionOnFire(where js.Value) error {
	p0, err := getCellXY(WhoThem, where)
	if err != nil {
		return err
	}
	cell := g.cell(WhoThem, p0)
	switch cell {
	case CellEmpty, CellSmoke:
		g.setCell(WhoThem, p0, CellMiss)
		where.Set("className", TDMiss)
		where.Set("innerText", EmojiMiss)
		return g.enemyFire()
	case CellShip:
		// continue after the switch.
		where.Set("className", TDHit)
		where.Set("innerText", EmojiHit)
	default:
		return fmt.Errorf("incorrect value of a cell(%s,%s) = %d", WhoThem, p0, cell)
	}

	// The ship has been hit.
	if _, err := g.shipHit(WhoThem, p0); err != nil {
		return err
	}
	if g.shipCount[WhoThem] == 0 {
		g.stop(WhoSelf)
	}
	return nil
}

func (g *Game) enemyFire() error {
	for fireAgain := true; fireAgain; {
		p := g.enemyPickCellToHit()
		td, err := g.tdCell(WhoSelf, p)
		if err != nil {
			return err
		}
		switch td.Get("className").String() {
		case TDEmpty:
			g.setCell(WhoSelf, p, CellMiss)
			td.Set("className", TDMiss)
			td.Set("innerText", EmojiMiss)
			fireAgain = false
		case TDShip:
			td.Set("className", TDHit)
			td.Set("innerText", EmojiHit)
			sunk, err := g.shipHit(WhoSelf, p)
			if err != nil {
				return err
			}
			if sunk {
				g.enemyPrevHit = xy{-1, -1}
			} else {
				g.enemyPrevHit = p
			}
			if g.shipCount[WhoSelf] == 0 {
				g.stop(WhoThem)
				return nil
			}
		}
	}
	return nil
}

func (g *Game) enemyPickCellToHit() xy {
	if g.cell(WhoSelf, g.enemyPrevHit) == CellHit {
		// Find boundaries of the hit cells of the ship.
		min, max := g.enemyPrevHit, g.enemyPrevHit
		for _, v := range allDirections {
			for p := g.enemyPrevHit.plus(v); g.cell(WhoSelf, p) == CellHit; p = p.plus(v) {
				if p.x < min.x {
					min.x = p.x
				}
				if p.y < min.y {
					min.y = p.y
				}
				if p.x > max.x {
					max.x = p.x
				}
				if p.y > max.y {
					max.y = p.y
				}
			}
		}
		preCells := make([]xy, 0, 4)
		switch {
		case min.x != max.x:
			preCells = append(preCells, min.plus(xy{-1, 0}), max.plus(xy{1, 0}))
		case min.y != max.y:
			preCells = append(preCells, min.plus(xy{0, -1}), max.plus(xy{0, 1}))
		default:
			preCells = append(preCells, min.plus(xy{-1, 0}), max.plus(xy{1, 0}), min.plus(xy{0, -1}), max.plus(xy{0, 1}))
		}
		cells := make([]xy, 0, 4)
		for _, c := range preCells {
			if g.cell(WhoSelf, c) != CellMiss {
				cells = append(cells, c)
			}
		}
		return cells[rand.Intn(len(cells))]
	}
	// We collect all cells that haven't been hit, i.e. either CellEmpty or CellShip.
	// We also count the number of adjacent empty/ship cells,
	// and only collect cells that have maximum number of such neighbors.
	neigh := 0
	var cells []xy
	for it := iter(xy{0, 0}, xy{GameSize, GameSize}); it.more(); it.next() {
		p := it.p
		c := g.cell(WhoSelf, p)
		if c >= CellSmoke {
			// It was hit before.
			continue
		}
		// Count adjacent cells that haven't been hit.
		n := 0
		for _, v := range allDirections {
			a := g.cell(WhoSelf, p.plus(v))
			if a < CellSmoke {
				n++
			}
		}
		switch {
		case n < neigh:
			// Too few non-hit adjacent cells, ignore this one.
		case n > neigh:
			// We've found a cell with more non-hit adjacent cells, reset the previous list.
			neigh = n
			cells = []xy{p}
		default:
			cells = append(cells, p)
		}
	}
	if len(cells) == 0 {
		// Cannot find a cell to hit -- checked all.
		return xy{-1, -1}
	}
	// Pick a random cell from the list.
	return cells[rand.Intn(len(cells))]
}

func (g *Game) stop(won Who) {
	g.cellOverListener.Remove(g.tableThem)
	g.cellOutListener.Remove(g.tableThem)
	g.cellClickListener.Remove(g.tableThem)
	name := "you have"
	if won == WhoThem {
		name = "AI has"
	}
	g.log("Game ended: %s won!!!", name)
}

func (g *Game) showStat(who Who) error {
	stat, err := g.getElementByID(fmt.Sprintf("%s-stat", who))
	if err != nil {
		return err
	}
	stat.Set("innerText", fmt.Sprintf("count:%d", g.shipCount[who]))
	return nil
}

// Note: the hit tdcell must be already set..
func (g *Game) shipHit(who Who, p0 xy) (bool, error) {
	g.setCell(who, p0, CellHit)
	g.shipCount[who] -= 1
	// Detect if the ship is sunk, and collect all ship cells.
	cells := []xy{p0}
	for _, v := range allDirections {
		for p, goon := p0.plus(v), true; goon; p = p.plus(v) {
			cell := g.cell(who, p)
			switch cell {
			case CellEmpty, CellMiss, CellSmoke:
				goon = false
			case CellShip:
				// The ship is not sunk yet.
				return false, g.showStat(who)
			case CellHit:
				cells = append(cells, p)
			default:
				return false, fmt.Errorf("incorrect value of cell(%s,%s) = %d", who, p, cell)
			}
		}
	}
	// The ship is sunk -- mark all its TD cells with TDDebris.
	for _, c := range cells {
		td, err := g.tdCell(who, c)
		if err != nil {
			return false, err
		}
		td.Set("className", TDDebris)
		td.Set("innerText", EmojiDebris)
	}
	// Also mark all empty cells near the ship with the smoke.
	for _, c := range cells {
		for _, v := range allDirections {
			s := c.plus(v)
			if g.cell(who, s) == CellEmpty {
				g.setCell(who, s, CellSmoke)
			}
		}
	}
	return true, g.showStat(who)
}

// DOM helpers.

func getClassList(obj js.Value) (js.Value, int) {
	clist := obj.Get("classList")
	if clist.Type() == js.TypeUndefined {
		return clist, 0
	}
	if len := clist.Get("length"); len.Type() == js.TypeNumber {
		return clist, len.Int()
	}
	return js.Undefined(), 0
}

func dbg(v js.Value) string {
	switch v.Type() {
	case js.TypeObject:
		sb := &strings.Builder{}
		sb.WriteString("<obj")
		if id := v.Get("id"); id.Type() != js.TypeUndefined && id.String() != "" {
			fmt.Fprintf(sb, " id=%s", id)
		}
		if typ := v.Get("type"); typ.Type() != js.TypeUndefined {
			fmt.Fprintf(sb, " type=%s", typ)
		}
		if clist, n := getClassList(v); n > 0 {
			fmt.Fprintf(sb, " cls=%s", clist.Get("value"))
		}
		sb.WriteString(">")
		return sb.String()
	default:
		return v.String()
	}
}

func dbga(a []js.Value) string {
	if len(a) == 0 {
		return "[]"
	}
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "[%d]{", len(a))
	for i, v := range a {
		if i != 0 {
			sb.WriteString(",")
		}
		fmt.Fprintf(sb, "%s", dbg(v))
	}
	sb.WriteString("}")
	return sb.String()
}

type EventListener struct {
	name string
	fn   js.Func
}

func newEventListener(evt string, fn func(js.Value, js.Value) any) *EventListener {
	return &EventListener{
		name: evt,
		fn: js.FuncOf(func(this js.Value, args []js.Value) any {
			if !this.Truthy() {
				fmt.Printf("event %q this is not truthy\n", evt)
				return nil
			}
			if len(args) != 1 {
				fmt.Printf("event %q len(args)=%d\n", evt, len(args))
				return nil
			}
			if !args[0].Truthy() {
				fmt.Printf("event %q arg[0] is not truthy\n", evt)
				return nil
			}
			fmt.Printf("Event %q called on %s evt=%s target=%s\n", evt, dbg(this), dbg(args[0]), dbg(args[0].Get("target")))
			return fn(this, args[0])
		}),
	}
}

func (e *EventListener) Add(elt js.Value) {
	fmt.Printf("Adding event listener %q to %s\n", e.name, dbg(elt))
	elt.Call("addEventListener", e.name, e.fn)
}

func (e *EventListener) Remove(elt js.Value) {
	fmt.Printf("Removing event listener %q from %s\n", e.name, dbg(elt))
	elt.Call("removeEventListener", e.name, e.fn)
}

func main() {
	// Export functions.
	// js.Global().Set("goTime", js.FuncOf(goTime))

	rand.Seed(time.Now().UnixNano())
	// Build the game, fill the grids.
	_, err := NewGame()
	if err != nil {
		fmt.Printf("Cannot create the game: %v\n", err)
		return
	}
	<-make(chan any)
}
