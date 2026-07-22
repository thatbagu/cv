package main

import (
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const steppeH = 16 // scene height in terminal rows

// ── Asset store (global, loaded once at startup) ──────────────────────────────

type steppeData struct {
	roofLeft, roofRight string
	bases               [6]string // open, open-clean, closed, closed-clean, light, light-clean
	stool, table        string
	fireFrames          [6]string
	horseMv             [3]string
	horseIdle, horseEat string
	dogMv               [2]string
	dogIdle             string
}

var steppeAssets *steppeData

func loadSteppeAssets(dir string) {
	load := func(name string) string {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return ""
		}
		return strings.TrimRight(string(b), "\n\r")
	}

	fpLines := strings.Split(load("fireplace.txt"), "\n")
	var fpBase string
	if len(fpLines) >= 3 {
		fpBase = strings.Join(fpLines[2:], "\n")
	}
	fireTops := [6]string{
		"     (    \n    ).    ",
		"     )    \n    (.    ",
		"    ( )   \n    )(    ",
		"    |(    \n    .)    ",
		"     (    \n    .)    ",
		"      )   \n    (.    ",
	}
	var fframes [6]string
	for i, t := range fireTops {
		fframes[i] = t + "\n" + fpBase
	}

	steppeAssets = &steppeData{
		roofLeft:  load("yurt_roof_left.txt"),
		roofRight: load("yurt_roof_right.txt"),
		bases: [6]string{
			load("yurt_base_open.txt"),
			load("yurt_base_open_clean.txt"),
			load("yurt_base_closed.txt"),
			load("yurt_base_closed_clean.txt"),
			load("yurt_base_light.txt"),
			load("yurt_base_light_clean.txt"),
		},
		stool:      load("stool.txt"),
		table:      load("table.txt"),
		fireFrames: fframes,
		horseMv: [3]string{
			load("horse_move_frame1.txt"),
			load("horse_move_frame2.txt"),
			load("horse_move_frame3.txt"),
		},
		horseIdle: load("horse_idle.txt"),
		horseEat:  load("horse_eat.txt"),
		dogMv: [2]string{
			load("dog_move_frame1.txt"),
			load("dog_move_frame2.txt"),
		},
		dogIdle: load("dog_idle.txt"),
	}
}

// ── Per-session scene state ───────────────────────────────────────────────────

type horseEnt struct {
	x, yBot  float64
	vx, vy   float64
	speed    float64
	state    string // "walk", "idle", "eat"
	frame    int
	frameCtr int
	stateCtr int
}

type dogEnt struct {
	x, yBot  float64
	vx, vy   float64
	speed    float64
	state    string
	frame    int
	frameCtr int
	stateCtr int
}

type yurtEnt struct {
	col, yBot int
	roofIdx   int // 0=left, 1=right
	baseIdx   int // 0..5
}

type fireEnt struct {
	col, yBot int
	frame     int
	frameCtr  int
}

type grassBlade struct {
	col, yBot int
	ch        rune
}

type stoolEnt struct {
	col, yBot int
}

type tableEnt struct {
	col, yBot int
}

type smokePuff struct {
	col, row float64
	char     rune
	driftX   float64
	age      int
	maxAge   int
}

type smokeEmitter struct {
	col, row int
	timer    int
}

type steppeScene struct {
	rng           *rand.Rand
	yurts         []yurtEnt
	fires         []fireEnt
	horses        []horseEnt
	dog           *dogEnt
	grass         []grassBlade
	stools        []stoolEnt
	tables        []tableEnt
	smokeEmitters []smokeEmitter
	smokePuffs    []smokePuff
	w             int
}

func newSteppeScene(w int) *steppeScene {
	if steppeAssets == nil || w < 30 {
		return nil
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	s := &steppeScene{rng: rng, w: w}

	// Yurts: 1 per ~70 cols, distributed evenly
	yurtCount := max(1, w/70)
	stride := w / yurtCount
	for i := 0; i < yurtCount; i++ {
		col := i*stride + rng.Intn(max(1, stride-36))
		s.yurts = append(s.yurts, yurtEnt{
			col:     max(0, col),
			yBot:    rng.Intn(3),
			roofIdx: i % 2,
			baseIdx: rng.Intn(6),
		})
	}

	// clearOfYurts returns true if col is far enough from all yurt origins.
	clearOfYurts := func(col int) bool {
		for _, y := range s.yurts {
			d := col - y.col
			if d < 0 {
				d = -d
			}
			if d < 35 {
				return false
			}
		}
		return true
	}

	// Fires: 1-3, placed away from yurts (up to 30 retries each)
	fireCount := 1 + rng.Intn(min(3, max(1, w/40)))
	for i := 0; i < fireCount; i++ {
		for try := 0; try < 30; try++ {
			col := 5 + rng.Intn(max(1, w-15))
			if clearOfYurts(col) {
				s.fires = append(s.fires, fireEnt{
					col:   col,
					yBot:  rng.Intn(3),
					frame: rng.Intn(6),
				})
				break
			}
		}
	}

	// clearOfFires returns true if col is far enough from all fire origins.
	// Fireplace art is 10 chars wide; 12-col margin avoids visual overlap.
	clearOfFires := func(col int) bool {
		for _, f := range s.fires {
			d := col - f.col
			if d < 0 {
				d = -d
			}
			if d < 12 {
				return false
			}
		}
		return true
	}

	// Stools and tables near each yurt, placed outside the yurt's bounding box
	// and away from campfires.  Yurt art is ~32 chars wide; col is its left edge.
	const yurtW = 32
	for _, y := range s.yurts {
		count := rng.Intn(3) // 0-2 stools per yurt
		for i := 0; i < count; i++ {
			side := 1
			if rng.Float64() < 0.5 {
				side = -1
			}
			var sc int
			if side == 1 {
				sc = y.col + yurtW + 2 + rng.Intn(8)
			} else {
				sc = y.col - 4 - rng.Intn(8)
			}
			if sc >= 0 && sc < w-3 && clearOfFires(sc) {
				s.stools = append(s.stools, stoolEnt{col: sc, yBot: y.yBot})
				if rng.Float64() < 0.35 {
					tc := sc + side*(3+rng.Intn(3))
					if tc >= 0 && tc < w-5 && clearOfFires(tc) {
						s.tables = append(s.tables, tableEnt{col: tc, yBot: y.yBot})
					}
				}
			}
		}
	}

	// Smoke emitters: one per yurt, positioned one row above the chimney.
	// The combined yurt art is 10 rows tall; art-line 0 (chimney) lands at
	// canvas row steppeH-10-yBot. The chimney center is at art col 15.
	for _, y := range s.yurts {
		chimneyRow := steppeH - 10 - y.yBot
		if chimneyRow >= 1 {
			s.smokeEmitters = append(s.smokeEmitters, smokeEmitter{
				col:   y.col + 15,
				row:   chimneyRow - 1,
				timer: rng.Intn(12), // stagger so chimneys don't puff in sync
			})
		}
	}

	// clearForGrass rejects a column that falls inside any static object's footprint.
	clearForGrass := func(col int) bool {
		for _, y := range s.yurts {
			if col >= y.col-1 && col < y.col+33 {
				return false
			}
		}
		for _, f := range s.fires {
			if col >= f.col-1 && col < f.col+11 {
				return false
			}
		}
		for _, sl := range s.stools {
			if col >= sl.col-1 && col < sl.col+3 {
				return false
			}
		}
		for _, tbl := range s.tables {
			if col >= tbl.col-1 && col < tbl.col+5 {
				return false
			}
		}
		return true
	}

	// Grass
	for i := 0; i < w/3; i++ {
		if rng.Float64() > 0.4 {
			continue
		}
		col := rng.Intn(w)
		if !clearForGrass(col) {
			continue
		}
		s.grass = append(s.grass, grassBlade{
			col:  col,
			yBot: rng.Intn(3),
			ch:   []rune{'W', 'w', '"', '\'', '^'}[rng.Intn(5)],
		})
	}

	// Horses: 1-4
	horseCount := 1 + rng.Intn(min(4, max(1, w/25)))
	for i := 0; i < horseCount; i++ {
		angle := rng.Float64() * math.Pi * 2
		s.horses = append(s.horses, horseEnt{
			x:        rng.Float64() * float64(w),
			yBot:     rng.Float64() * 4,
			vx:       math.Cos(angle),
			vy:       math.Sin(angle) * 0.25,
			speed:    12 + rng.Float64()*15,
			state:    "walk",
			stateCtr: 50 + rng.Intn(80),
		})
	}

	// Dog: 50% chance
	if rng.Float64() < 0.5 {
		angle := rng.Float64() * math.Pi * 2
		s.dog = &dogEnt{
			x:        rng.Float64() * float64(w),
			yBot:     rng.Float64() * 4,
			vx:       math.Cos(angle),
			vy:       math.Sin(angle) * 0.25,
			speed:    18 + rng.Float64()*15,
			state:    "walk",
			stateCtr: 40 + rng.Intn(60),
		}
	}

	return s
}

// tickSteppeScene advances the simulation by one frame (called at 24 fps).
func tickSteppeScene(s *steppeScene) {
	if s == nil || steppeAssets == nil {
		return
	}
	const dt = 1.0 / 24.0
	const maxY = 4.0
	w := float64(s.w)

	// Fire animation (~8 fps)
	for i := range s.fires {
		f := &s.fires[i]
		f.frameCtr++
		if f.frameCtr >= 3 {
			f.frameCtr = 0
			f.frame = (f.frame + 1) % 6
		}
	}

	// Smoke: age and rise existing puffs, cull dead ones
	smokeChars := []rune{'~', ',', '.', '\''}
	live := s.smokePuffs[:0]
	for _, p := range s.smokePuffs {
		p.row -= 0.15
		p.col += p.driftX
		p.age++
		if p.row >= 0 && p.age < p.maxAge {
			live = append(live, p)
		}
	}
	s.smokePuffs = live

	// Smoke: emit new puffs from each chimney
	for i := range s.smokeEmitters {
		e := &s.smokeEmitters[i]
		e.timer++
		if e.timer >= 8+s.rng.Intn(8) {
			e.timer = 0
			jitter := (s.rng.Float64() - 0.5) * 1.5
			s.smokePuffs = append(s.smokePuffs, smokePuff{
				col:    float64(e.col) + jitter,
				row:    float64(e.row),
				char:   smokeChars[s.rng.Intn(len(smokeChars))],
				driftX: (s.rng.Float64()*2 - 1) * 0.08,
				maxAge: 20 + s.rng.Intn(20),
			})
		}
	}

	// Horses
	for i := range s.horses {
		h := &s.horses[i]
		h.stateCtr--
		if h.stateCtr <= 0 {
			h.stateCtr = 50 + s.rng.Intn(100)
			r := s.rng.Float64()
			switch {
			case r < 0.55:
				h.state = "walk"
				angle := s.rng.Float64() * math.Pi * 2
				h.vx = math.Cos(angle)
				h.vy = math.Sin(angle) * 0.25
			case r < 0.8:
				h.state = "idle"
			default:
				h.state = "eat"
			}
		}
		if h.state == "walk" {
			h.x += h.vx * h.speed * dt
			h.yBot += h.vy * h.speed * dt
			if h.x > w+12 {
				h.x = -12
			}
			if h.x < -12 {
				h.x = w + 12
			}
			if h.yBot < 0 {
				h.yBot = 0
				h.vy = -h.vy
			}
			if h.yBot > maxY {
				h.yBot = maxY
				h.vy = -h.vy
			}
			h.frameCtr++
			if h.frameCtr >= 5 {
				h.frameCtr = 0
				h.frame = (h.frame + 1) % 3
			}
		}
	}

	// Dog
	if s.dog != nil {
		d := s.dog
		d.stateCtr--
		if d.stateCtr <= 0 {
			d.stateCtr = 40 + s.rng.Intn(80)
			if s.rng.Float64() < 0.65 {
				d.state = "walk"
				angle := s.rng.Float64() * math.Pi * 2
				d.vx = math.Cos(angle)
				d.vy = math.Sin(angle) * 0.25
			} else {
				d.state = "idle"
			}
		}
		if d.state == "walk" {
			d.x += d.vx * d.speed * dt
			d.yBot += d.vy * d.speed * dt
			if d.x > w+6 {
				d.x = -6
			}
			if d.x < -6 {
				d.x = w + 6
			}
			if d.yBot < 0 {
				d.yBot = 0
				d.vy = -d.vy
			}
			if d.yBot > maxY {
				d.yBot = maxY
				d.vy = -d.vy
			}
			d.frameCtr++
			if d.frameCtr >= 4 {
				d.frameCtr = 0
				d.frame = (d.frame + 1) % 2
			}
		}
		// Horse repulsion
		for i := range s.horses {
			h := &s.horses[i]
			dx := h.x - d.x
			dy := h.yBot - d.yBot
			if (dx/20)*(dx/20)+(dy/2)*(dy/2) < 1 {
				if h.state != "walk" {
					h.state = "walk"
					h.stateCtr = 40
				}
				l := math.Sqrt(dx*dx + dy*dy)
				if l > 0 {
					h.vx = h.vx*0.7 + (dx/l)*0.3
					h.vy = h.vy*0.7 + (dy/l)*0.3
				}
			}
		}
	}
}

// renderSteppeScene composites all scene elements onto a character canvas.
func renderSteppeScene(s *steppeScene, st styles) string {
	if s == nil || steppeAssets == nil {
		return ""
	}
	w := s.w
	h := steppeH

	// Canvas: h rows × w cols, ' ' = empty
	canvas := make([][]rune, h)
	for i := range canvas {
		canvas[i] = make([]rune, w)
		for j := range canvas[i] {
			canvas[i][j] = ' '
		}
	}

	// mirrorRune swaps paired/directional ASCII characters for horizontal mirroring.
	mirrorRune := func(r rune) rune {
		switch r {
		case '/':
			return '\\'
		case '\\':
			return '/'
		case '(':
			return ')'
		case ')':
			return '('
		case '[':
			return ']'
		case ']':
			return '['
		case '{':
			return '}'
		case '}':
			return '{'
		case '<':
			return '>'
		case '>':
			return '<'
		}
		return r
	}

	// blit places art with its bottom-left at (col, yBot).
	// yBot=0 → bottom row of art sits at canvas row h-1 (ground).
	// Non-space runes only (spaces are transparent).
	blit := func(art string, col, yBot int, flipH bool) {
		lines := strings.Split(strings.TrimRight(art, "\n"), "\n")
		n := len(lines)

		// When flipping, all lines must share the same width so that shorter
		// lines (e.g. leg rows) don't shift left relative to longer lines.
		maxW := 0
		if flipH {
			for _, line := range lines {
				if l := len([]rune(line)); l > maxW {
					maxW = l
				}
			}
		}

		bottomRow := h - 1 - yBot
		for li, line := range lines {
			row := bottomRow - (n - 1 - li)
			if row < 0 || row >= h {
				continue
			}
			runes := []rune(line)
			if flipH {
				// Pad to maxW so every line flips around the same right edge.
				for len(runes) < maxW {
					runes = append(runes, ' ')
				}
				// Reverse order, then swap directional chars.
				for l, r := 0, len(runes)-1; l < r; l, r = l+1, r-1 {
					runes[l], runes[r] = runes[r], runes[l]
				}
				for i, r := range runes {
					runes[i] = mirrorRune(r)
				}
			}
			for j, ch := range runes {
				x := col + j
				if x < 0 || x >= w {
					continue
				}
				if ch != ' ' {
					canvas[row][x] = ch
				}
			}
		}
	}

	// Collect all draw operations and depth-sort (painter's algorithm).
	type blitOp struct {
		art   string
		col   int
		yBot  float64
		flipH bool
	}
	var ops []blitOp

	for _, g := range s.grass {
		ops = append(ops, blitOp{string(g.ch), g.col, float64(g.yBot), false})
	}
	for _, y := range s.yurts {
		roof := steppeAssets.roofLeft
		if y.roofIdx == 1 {
			roof = steppeAssets.roofRight
		}
		ops = append(ops, blitOp{roof + "\n" + steppeAssets.bases[y.baseIdx], y.col, float64(y.yBot), false})
	}
	for _, f := range s.fires {
		ops = append(ops, blitOp{steppeAssets.fireFrames[f.frame], f.col, float64(f.yBot), false})
	}
	for _, sl := range s.stools {
		ops = append(ops, blitOp{steppeAssets.stool, sl.col, float64(sl.yBot), false})
	}
	for _, tbl := range s.tables {
		ops = append(ops, blitOp{steppeAssets.table, tbl.col, float64(tbl.yBot), false})
	}
	for _, h := range s.horses {
		var art string
		switch h.state {
		case "idle":
			art = steppeAssets.horseIdle
		case "eat":
			art = steppeAssets.horseEat
		default:
			art = steppeAssets.horseMv[h.frame%3]
		}
		ops = append(ops, blitOp{art, int(h.x), h.yBot, h.vx < 0})
	}
	if s.dog != nil {
		d := s.dog
		var art string
		if d.state == "idle" {
			art = steppeAssets.dogIdle
		} else {
			art = steppeAssets.dogMv[d.frame%2]
		}
		ops = append(ops, blitOp{art, int(d.x), d.yBot, d.vx > 0})
	}

	// Sort farthest-first (lower yBot = farther from viewer)
	sort.Slice(ops, func(i, j int) bool { return ops[i].yBot < ops[j].yBot })

	for _, op := range ops {
		blit(op.art, op.col, int(op.yBot), op.flipH)
	}

	// Smoke is rendered after the painter's algorithm — it floats above everything.
	for _, p := range s.smokePuffs {
		r := int(p.row)
		c := int(p.col)
		if r >= 0 && r < h && c >= 0 && c < w {
			canvas[r][c] = p.char
		}
	}

	// Render canvas with accent color.
	var sb strings.Builder
	for _, row := range canvas {
		line := strings.TrimRight(string(row), " ")
		if line != "" {
			sb.WriteString(st.peach.Render(line))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
