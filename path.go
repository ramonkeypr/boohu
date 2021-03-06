package main

func (d *dungeon) FreeNeighbors(pos position) []position {
	neighbors := [8]position{pos.E(), pos.W(), pos.N(), pos.S(), pos.NE(), pos.NW(), pos.SE(), pos.SW()}
	freeNeighbors := []position{}
	for _, npos := range neighbors {
		if d.Valid(npos) && d.Cell(npos).T != WallCell {
			freeNeighbors = append(freeNeighbors, npos)
		}
	}
	return freeNeighbors
}

func (d *dungeon) CardinalFreeNeighbors(pos position) []position {
	neighbors := [4]position{pos.E(), pos.W(), pos.N(), pos.S()}
	freeNeighbors := []position{}
	for _, npos := range neighbors {
		if d.Valid(npos) && d.Cell(npos).T != WallCell {
			freeNeighbors = append(freeNeighbors, npos)
		}
	}
	return freeNeighbors
}

type playerPath struct {
	game *game
}

func (pp *playerPath) Neighbors(pos position) []position {
	m := pp.game.Dungeon
	var neighbors []position
	if pp.game.Player.HasStatus(StatusConfusion) {
		neighbors = m.CardinalFreeNeighbors(pos)
	} else {
		neighbors = m.FreeNeighbors(pos)
	}
	freeNeighbors := []position{}
	for _, npos := range neighbors {
		if m.Cell(npos).Explored && !pp.game.UnknownDig[npos] && !pp.game.ExclusionsMap[npos] {
			freeNeighbors = append(freeNeighbors, npos)
		}
	}
	return freeNeighbors
}

func (pp *playerPath) Cost(from, to position) int {
	return 1
}

func (pp *playerPath) Estimation(from, to position) int {
	return from.Distance(to)
}

type normalPath struct {
	game *game
}

func (np *normalPath) Neighbors(pos position) []position {
	if np.game.Player.HasStatus(StatusConfusion) {
		return np.game.Dungeon.CardinalFreeNeighbors(pos)
	}
	return np.game.Dungeon.FreeNeighbors(pos)
}

func (np *normalPath) Cost(from, to position) int {
	return 1
}

type autoexplorePath struct {
	game *game
}

func (ap *autoexplorePath) Neighbors(pos position) []position {
	if ap.game.ExclusionsMap[pos] {
		return nil
	}
	var neighbors []position
	if ap.game.Player.HasStatus(StatusConfusion) {
		neighbors = ap.game.Dungeon.CardinalFreeNeighbors(pos)
	} else {
		neighbors = ap.game.Dungeon.FreeNeighbors(pos)
	}
	var suitableNeighbors []position
	for _, pos := range neighbors {
		if !ap.game.ExclusionsMap[pos] {
			suitableNeighbors = append(suitableNeighbors, pos)
		}
	}
	return suitableNeighbors
}

func (ap *autoexplorePath) Cost(from, to position) int {
	return 1
}

type monPath struct {
	game    *game
	monster *monster
	wall    bool
}

func (mp *monPath) Neighbors(pos position) []position {
	if mp.monster.Status(MonsConfused) {
		if mp.wall {
			return mp.game.Dungeon.CardinalNeighbors(pos)
		}
		return mp.game.Dungeon.CardinalFreeNeighbors(pos)
	}
	if mp.wall {
		return mp.game.Dungeon.Neighbors(pos)
	}
	return mp.game.Dungeon.FreeNeighbors(pos)
}

func (mp *monPath) Cost(from, to position) int {
	g := mp.game
	mons, _ := g.MonsterAt(to)
	if !mons.Exists() {
		if mp.wall && g.Dungeon.Cell(to).T == WallCell && mp.monster.State != Hunting {
			return 6
		}
		return 1
	}
	return 4
}

func (mp *monPath) Estimation(from, to position) int {
	return from.Distance(to)
}

func (m *monster) APath(g *game, from, to position) []position {
	mp := &monPath{game: g, monster: m}
	if m.Kind == MonsEarthDragon {
		mp.wall = true
	}
	path, _, found := AstarPath(mp, from, to)
	if !found {
		return nil
	}
	return path
}

func (g *game) PlayerPath(from, to position) []position {
	pp := &playerPath{game: g}
	path, _, found := AstarPath(pp, from, to)
	if !found {
		return nil
	}
	return path
}
