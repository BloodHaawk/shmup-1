package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten"
)

const (
	squareSize = 8
	hitBoxSize = 2
	mvtSpeed   = 3

	maxPlayerBullets  = 1000
	playerBulletSize  = 3
	playerBulletSpeed = 5
	playerBulletFreq  = 60
	baseBulletSpread  = 100 //degrees
	focusBulletSpread = 90  //degrees
	bulletStreams     = 9
)

// Player struct
type player struct {
	skin          sprite
	hitBox        sprite
	bullets       []bullet
	bulletSkin    *ebiten.Image
	bulletSprite  sprite
	lastShotFrame int
	isFocus       bool
}

func (p *player) update(screen *ebiten.Image) {
	p.move(mvtSpeed)

	// Draw the square and update the position from keyboard input
	drawSprite(screen, p.skin)

	// Show the hitBox in red when pressing Shift
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		drawSprite(screen, p.hitBox)
		p.isFocus = true
	} else {
		p.isFocus = false
	}

	// Shoot a bullet with Z key
	if ebiten.IsKeyPressed(ebiten.KeyZ) {
		if p.isFocus {
			p.shootBullet(playerBulletFreq, bulletStreams, focusBulletSpread)
		} else {
			p.shootBullet(playerBulletFreq, bulletStreams, baseBulletSpread)
		}
	}

	for i := range p.bullets {
		if p.bullets[i].isOnScreen {
			p.bulletSprite.opts.GeoM.Reset()
			p.bulletSprite.opts.GeoM.Translate(p.bullets[i].x, p.bullets[i].y)
			drawSprite(screen, p.bulletSprite)
			p.bullets[i].move(playerBulletSpeed, playerBulletSize)
		}
	}
}

// Move a sprite from keyboard inputs (use Shift to slow down)
func (p *player) move(speed float64) {
	// Use Shift to slow down
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		speed /= 2
	}
	var tx, ty float64

	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		tx = 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		tx = -1
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		ty = 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		ty = -1
	}

	if r := math.Sqrt(tx*tx + ty*ty); r != 0 {
		tx = tx / r * speed
		ty = ty / r * speed

		tx = math.Max(0, p.hitBox.x()+tx) - p.hitBox.x()
		tx = math.Min(windowWidth-hitBoxSize, p.hitBox.x()+tx) - p.hitBox.x()

		ty = math.Max(0, p.hitBox.y()+ty) - p.hitBox.y()
		ty = math.Min(windowHeight-hitBoxSize, p.hitBox.y()+ty) - p.hitBox.y()
	}

	p.skin.opts.GeoM.Translate(tx, ty)
	p.hitBox.opts.GeoM.Translate(tx, ty)

	return
}

func (p *player) shootBullet(freq int, n int, spreadDeg float64) {

	if frameCounter-p.lastShotFrame >= ebiten.MaxTPS()/playerBulletFreq {
		indices := findNFirsts(p.bullets, n, func(b bullet) bool { return !b.isOnScreen })

		if len(indices) == n {
			for i := 0; i < n; i++ {
				angleDeg := -spreadDeg/2 + float64(i)*spreadDeg/float64(n-1)
				p.bullets[indices[i]].x = p.hitBox.x() + 15*math.Sin(angleDeg*math.Pi/180) + (hitBoxSize-playerBulletSize)/2
				p.bullets[indices[i]].y = p.hitBox.y() - 15*math.Cos(angleDeg*math.Pi/180) + (hitBoxSize-playerBulletSize)/2
				if p.isFocus {
					p.bullets[indices[i]].vx = 0
					p.bullets[indices[i]].vy = -1
				} else {
					p.bullets[indices[i]].vx = math.Sin(angleDeg * math.Pi / 180)
					p.bullets[indices[i]].vy = -math.Cos(angleDeg * math.Pi / 180)
				}
				p.bullets[indices[i]].isOnScreen = true
			}
		}
		p.lastShotFrame = frameCounter
	}

}

func initPlayer() player {
	var p player
	var errH, errS, errB error
	p.hitBox.image, errH = ebiten.NewImage(hitBoxSize, hitBoxSize, ebiten.FilterNearest)
	p.skin.image, errS = ebiten.NewImage(squareSize, squareSize, ebiten.FilterNearest)
	logError(errH)
	logError(errS)

	p.hitBox.image.Fill(color.RGBA{255, 0, 0, 255})
	p.skin.image.Fill(color.White)
	p.skin.opts.GeoM.Translate((hitBoxSize-squareSize)/2, (hitBoxSize-squareSize)/2)

	// Start at middle of screen
	p.hitBox.opts.GeoM.Translate((windowWidth-hitBoxSize)/2, (windowHeight-hitBoxSize)/2)
	p.skin.opts.GeoM.Translate((windowWidth-hitBoxSize)/2, (windowHeight-hitBoxSize)/2)

	p.bulletSkin, errB = ebiten.NewImage(playerBulletSize, playerBulletSize, ebiten.FilterNearest)
	logError(errB)
	p.bulletSkin.Fill(color.White)
	p.bulletSprite = sprite{p.bulletSkin, ebiten.DrawImageOptions{}}

	p.bullets = make([]bullet, maxPlayerBullets)

	return p
}
