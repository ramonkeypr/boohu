package main

import (
	"container/heap"
	"errors"
	"fmt"
)

type rod int

const (
	RodDigging rod = iota
	RodBlink
	RodTeleportOther
	RodLightningBolt
	RodFireball
	RodFog
	RodShatter
	// below unimplemented
	RodConfusingClouds
	RodFear
	RodFreezingClouds
)

func (r rod) Letter() rune {
	return '/'
}

func (r rod) Rare() bool {
	switch r {
	case RodDigging, RodTeleportOther, RodShatter:
		return true
	default:
		return false
	}
}

func (r rod) String() string {
	var text string
	switch r {
	case RodDigging:
		text = "rod of digging"
	case RodBlink:
		text = "rod of blinking"
	case RodTeleportOther:
		text = "rod of teleport other"
	case RodFog:
		text = "rod of fog"
	case RodFear:
		text = "rod of fear"
	case RodFireball:
		text = "rod of fireball"
	case RodLightningBolt:
		text = "rod of lightning bolt"
	case RodShatter:
		text = "rod of shatter"
	case RodFreezingClouds:
		text = "rod of freezing clouds"
	case RodConfusingClouds:
		text = "rod of confusing clouds"
	}
	return text
}

func (r rod) Desc() string {
	var text string
	switch r {
	case RodDigging:
		text = "digs through walls."
	case RodBlink:
		text = "makes you blink away within your line of sight."
	case RodTeleportOther:
		text = "teleports away one of your foes."
	case RodFog:
		text = "creates a dense fog that reduces your (and monster's) line of sight."
	case RodFireball:
		text = "throws a 1-radius fireball at your foes."
	case RodLightningBolt:
		text = "throws a lightning bolt through one or more ennemies."
	case RodShatter:
		text = "induces an explosion around a wall. The wall can disintegrate."
	case RodFear:
		text = "TODO"
	case RodFreezingClouds:
		text = "TODO"
	case RodConfusingClouds:
		text = "TODO"
	}
	return fmt.Sprintf("The %s %s", r, text)
}

type rodProps struct {
	Charge int
}

func (r rod) MaxCharge() (charges int) {
	switch r {
	case RodBlink:
		charges = 5
	case RodTeleportOther, RodDigging:
		charges = 3
	default:
		charges = 4
	}
	return charges
}

func (r rod) Rate() int {
	rate := r.MaxCharge() - 2
	if rate < 1 {
		rate = 1
	}
	return rate
}

func (r rod) MPCost() (mp int) {
	switch r {
	case RodBlink:
		mp = 3
	case RodTeleportOther, RodDigging, RodShatter:
		mp = 5
	default:
		mp = 4
	}
	return mp
}

func (r rod) Use(g *game, ev event) error {
	rods := g.Player.Rods
	if rods[r].Charge <= 0 {
		return errors.New("No charges remaining on this rod.")
	}
	if r.MPCost() > g.Player.MP {
		return errors.New("Not enough magic points for using this rod.")
	}
	var err error
	switch r {
	case RodBlink:
		err = g.EvokeRodBlink(ev)
	case RodTeleportOther:
		err = g.EvokeRodTeleportOther(ev)
	case RodLightningBolt:
		err = g.EvokeRodLightningBolt(ev)
	case RodFireball:
		err = g.EvokeRodFireball(ev)
	case RodFog:
		err = g.EvokeRodFog(ev)
	case RodDigging:
		err = g.EvokeRodDigging(ev)
	case RodShatter:
		err = g.EvokeRodShatter(ev)
	}

	if err != nil {
		return err
	}
	rods[r].Charge--
	g.Player.MP -= r.MPCost()
	g.StoryPrintf("You evoked your %s.", r)
	g.FairAction()
	ev.Renew(g, 7)
	return nil
}

func (g *game) EvokeRodBlink(ev event) error {
	if g.Player.HasStatus(StatusLignification) {
		return errors.New("You cannot blink while lignified.")
	}
	g.Blink(ev)
	return nil
}

func (g *game) Blink(ev event) {
	if g.Player.HasStatus(StatusLignification) {
		return
	}
	losPos := []position{}
	for pos, b := range g.Player.LOS {
		if !b {
			continue
		}
		if g.Dungeon.Cell(pos).T != FreeCell {
			continue
		}
		mons, _ := g.MonsterAt(pos)
		if mons.Exists() {
			continue
		}
		losPos = append(losPos, pos)
	}
	if len(losPos) == 0 {
		// should not happen
		g.Print("You could not blink.")
		return
	}
	npos := losPos[RandInt(len(losPos))]
	if npos.Distance(g.Player.Pos) <= 3 {
		// Give close cells less chance to make blinking more useful
		npos = losPos[RandInt(len(losPos))]
	}
	g.Player.Pos = npos
	g.Print("You blink away.")
	g.ComputeLOS()
	g.MakeMonstersAware()
}

func (g *game) EvokeRodTeleportOther(ev event) error {
	if !g.ui.ChooseTarget(g, &chooser{}) {
		return errors.New("Ok, then.")
	}
	mons, _ := g.MonsterAt(g.Player.Target)
	if mons == nil {
		// should not happen (done in the targeter)
		return errors.New("You must target a monster for using this rod.")
	}
	pos := mons.Pos
	i := 0
	count := 0
	for {
		count++
		if count > 1000 {
			panic("TeleportOther")
		}
		pos = g.FreeCell()
		if pos.Distance(mons.Pos) < 15 && i < 1000 {
			i++
			continue
		}
		break
	}

	switch mons.State {
	case Hunting:
		mons.State = Wandering
		// TODO: change the target?
	case Resting, Wandering:
		mons.State = Wandering
		mons.Target = mons.Pos
	}
	mons.Pos = pos
	g.Printf("The %s teleports away.", mons.Kind)
	return nil
}

func (g *game) EvokeRodLightningBolt(ev event) error {
	if !g.ui.ChooseTarget(g, &chooser{}) {
		return errors.New("Ok, then.")
	}
	ray := g.Ray(g.Player.Target)
	g.Print("A lightning bolt emerges straight from the rod.")
	for _, pos := range ray {
		mons, _ := g.MonsterAt(pos)
		if mons == nil {
			continue
		}
		mons.HP -= RandInt(21)
		if mons.HP <= 0 {
			g.Printf("%s is killed by the bolt.", Indefinite(mons.Kind.String(), true))
			g.KillStats(mons)
		}
		g.MakeNoise(12, mons.Pos)
		mons.MakeHuntIfHurt(g)
	}
	return nil
}

func (g *game) EvokeRodFireball(ev event) error {
	if !g.ui.ChooseTarget(g, &chooser{area: true, minDist: true}) {
		return errors.New("Ok, then.")
	}
	neighbors := g.Dungeon.FreeNeighbors(g.Player.Target)
	g.Print("A fireball emerges straight from the rod.")
	for _, pos := range append(neighbors, g.Player.Target) {
		mons, _ := g.MonsterAt(pos)
		if mons == nil {
			continue
		}
		mons.HP -= RandInt(21)
		if mons.HP <= 0 {
			g.Printf("%s is killed by the fireball.", Indefinite(mons.Kind.String(), true))
			g.KillStats(mons)
		}
		g.MakeNoise(12, mons.Pos)
		mons.MakeHuntIfHurt(g)
	}
	return nil
}

type cloud int

const (
	CloudFog cloud = iota
)

func (g *game) EvokeRodFog(ev event) error {
	dij := &normalPath{game: g}
	nm := Dijkstra(dij, []position{g.Player.Pos}, 3)
	for pos := range nm {
		_, ok := g.Clouds[pos]
		if !ok {
			g.Clouds[pos] = CloudFog
			heap.Push(g.Events, &cloudEvent{ERank: ev.Rank() + 100 + RandInt(100), EAction: CloudEnd, Pos: pos})
		}
	}
	g.ComputeLOS()
	g.Print("You are surrounded by a dense fog.")
	return nil
}

func (g *game) EvokeRodDigging(ev event) error {
	if !g.ui.ChooseTarget(g, &wallChooser{}) {
		return errors.New("Ok, then.")
	}
	pos := g.Player.Target
	for i := 0; i < 3; i++ {
		g.Dungeon.SetCell(pos, FreeCell)
		g.MakeNoise(17, pos)
		pos = pos.To(pos.Dir(g.Player.Pos))
		if !g.Player.LOS[pos] {
			g.UnknownDig[pos] = true
		}
		if !g.Dungeon.Valid(pos) || g.Dungeon.Cell(pos).T != WallCell {
			break
		}
	}
	g.Print("You see the wall disintegrate with a crash.")
	g.ComputeLOS()
	g.MakeMonstersAware()
	return nil
}

func (g *game) EvokeRodShatter(ev event) error {
	if !g.ui.ChooseTarget(g, &wallChooser{minDist: true}) {
		return errors.New("Ok, then.")
	}
	neighbors := g.Dungeon.FreeNeighbors(g.Player.Target)
	if RandInt(2) == 0 {
		g.Dungeon.SetCell(g.Player.Target, FreeCell)
		g.ComputeLOS()
		g.MakeMonstersAware()
		g.MakeNoise(19, g.Player.Target)
		g.Print("You see the wall disappear in a noisy explosion.")
	} else {
		g.MakeNoise(15, g.Player.Target)
		g.Print("You see an explosion around the wall.")
	}
	for _, pos := range neighbors {
		mons, _ := g.MonsterAt(pos)
		if mons == nil {
			continue
		}
		mons.HP -= RandInt(30)
		if mons.HP <= 0 {
			g.Printf("%s is killed by the explosion.", Indefinite(mons.Kind.String(), true))
			g.KillStats(mons)
		}
		g.MakeNoise(12, mons.Pos)
		mons.MakeHuntIfHurt(g)
	}
	return nil
}

func (g *game) GeneratedRodsCount() int {
	count := 0
	for _, b := range g.GeneratedRods {
		if b {
			count++
		}
	}
	return count
}

func (g *game) GenerateRod() {
	count := 0
	for {
		count++
		if count > 1000 {
			panic("GenerateRod")
		}
		pos := g.FreeCellForStatic()
		r := rod(RandInt(int(RodShatter) + 1))
		if r.Rare() {
			r = rod(RandInt(int(RodShatter) + 1))
		}
		if g.Player.Rods[r] == nil && !g.GeneratedRods[r] {
			g.GeneratedRods[r] = true
			g.Rods[pos] = r
			return
		}
	}
}
