package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten"
)

type pattern struct {
	bullets []bullet
	aimLine float64 // Signed angle in degrees between aimline and vertical reference
	isAimed bool    // Switches between static and aimed patterns

	opts patternOpts
}

type patternOpts struct {
	bulletSkin   *ebiten.Image
	bulletSprite sprite

	bulletSize        int
	bulletSpeed       float64
	bulletFreq        int
	bulletSpread      float64
	bulletStreams     int
	bulletSpawnOffset float64

	bulletColor color.Color
}

func (p *pattern) updateBullets(screen *ebiten.Image, hb hitBox, pl *player) {
	for i := range p.bullets {
		if p.bullets[i].isOnScreen {
			p.opts.bulletSprite.opts.GeoM.Reset()
			p.opts.bulletSprite.opts.GeoM.Translate(p.bullets[i].x, p.bullets[i].y)
			drawSprite(screen, p.opts.bulletSprite)
			p.bullets[i].move(p.opts.bulletSpeed, float64(p.opts.bulletSize))
			if collision(pl, &p.bullets[i]) {
				p.bullets[i].isOnScreen = false
				screen.Fill(color.RGBA{255, 0, 0, 255})
			}
		}
	}

	p.aimLine = p.getAimLine(hb, pl)
}

func (p *pattern) getAimLine(hb hitBox, pl *player) (alpha float64) {
	alpha = math.Atan2(pl.posX()-hb.x, pl.posY()-hb.y)
	return
}

func (p *pattern) spawn(hb hitBox) {
	if indices := findNFirsts(p.bullets, p.opts.bulletStreams, func(b bullet) bool { return !b.isOnScreen }); len(indices) == p.opts.bulletStreams {
		for i := range indices {
			angleDeg := -p.opts.bulletSpread/2 + float64(i)*p.opts.bulletSpread/float64(p.opts.bulletStreams-1)
			p.bullets[indices[i]].x = hb.x + p.opts.bulletSpawnOffset*math.Sin(angleDeg*math.Pi/180) + (hb.xSize-float64(p.opts.bulletSize))/2
			p.bullets[indices[i]].y = hb.y + p.opts.bulletSpawnOffset*math.Cos(angleDeg*math.Pi/180) + (hb.ySize-float64(p.opts.bulletSize))/2
			p.bullets[indices[i]].vx = math.Sin(angleDeg * math.Pi / 180)
			p.bullets[indices[i]].vy = math.Cos(angleDeg * math.Pi / 180)
			p.bullets[indices[i]].isOnScreen = true
		}

		if p.isAimed {
			for i := range indices {
				p.bullets[indices[i]].vx = p.bullets[indices[i]].vx*math.Cos(p.aimLine) + p.bullets[indices[i]].vy*math.Sin(p.aimLine)
				p.bullets[indices[i]].vy = -p.bullets[indices[i]].vx*math.Sin(p.aimLine) + p.bullets[indices[i]].vy*math.Cos(p.aimLine)
			}
		}
	}

}

func initPattern(isAimed bool, opts patternOpts) pattern {

	bullets := make([]bullet, maxBullets)
	var aimLine float64 = 0

	return pattern{
		bullets,
		aimLine,
		isAimed,

		opts,
	}

}
