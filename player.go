package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten"

	"github.com/bloodhaawk/shmup-1/collision"
	"github.com/bloodhaawk/shmup-1/controlmap"
	"github.com/bloodhaawk/shmup-1/utils"
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

	vx, vy float64

	skinSize   int
	hitBoxSize int
	mvtSpeed   float64

	bulletSize        int
	bulletSpeed       float64
	bulletFreq        int
	baseBulletSpread  float64 //degrees
	focusBulletSpread float64 //degrees
	bulletStreams     int
	bulletSpawnOffset float64
}

// Implement collisionBox interface

func (p *player) PosX() float64 {
	return p.hitBox.x()
}
func (p *player) PosY() float64 {
	return p.hitBox.y()
}
func (p *player) VX() float64 {
	return p.vx
}
func (p *player) VY() float64 {
	return p.vy
}
func (p *player) SizeX() float64 {
	return float64(p.hitBoxSize)
}
func (p *player) SizeY() float64 {
	return float64(p.hitBoxSize)
}

func (p *player) centreX() float64 {
	return p.hitBox.x() + float64(p.hitBoxSize)/2
}
func (p *player) centreY() float64 {
	return p.hitBox.y() + float64(p.hitBoxSize)/2
}

func (p *player) update(screen *ebiten.Image) {
	p.move(p.mvtSpeed)

	// Draw the square and update the position from keyboard input
	drawSprite(screen, p.skin)

	// Show the hitBox in red when pressing focus
	if ebiten.IsKeyPressed(controlmap.KeyMap[keyConfig["focus"]]) || ebiten.IsGamepadButtonPressed(gamepadID, controlmap.ButtonMap[buttonConfig["focus"]]) {
		drawSprite(screen, p.hitBox)
		p.isFocus = true
	} else {
		p.isFocus = false
	}

}

// Update player bullets
func (p *player) updateBullets(screen *ebiten.Image, e []enemy) {
	if ebiten.IsKeyPressed(controlmap.KeyMap[keyConfig["shoot"]]) || ebiten.IsGamepadButtonPressed(gamepadID, controlmap.ButtonMap[buttonConfig["shoot"]]) {
		if p.isFocus {
			p.shootBullets(p.bulletFreq, p.bulletStreams, p.focusBulletSpread)
		} else {
			p.shootBullets(p.bulletFreq, p.bulletStreams, p.baseBulletSpread)
		}
	}

	for i := range p.bullets {
		if p.bullets[i].isOnScreen {
			p.bulletSprite.opts.GeoM.Reset()
			p.bulletSprite.opts.GeoM.Translate(p.bullets[i].x, p.bullets[i].y)
			drawSprite(screen, p.bulletSprite)
			for j := range e {
				if collision.Collision(&p.bullets[i], &e[j]) {
					p.bullets[i].isOnScreen = false
					break
				}
			}
			p.bullets[i].move(p.bulletSpeed, float64(p.bulletSize))
		}
	}
}

// Move a sprite from keyboard inputs (use Shift to slow down)
func (p *player) move(speed float64) {
	// Use Shift to slow down
	if ebiten.IsKeyPressed(controlmap.KeyMap[keyConfig["focus"]]) || ebiten.IsGamepadButtonPressed(gamepadID, controlmap.ButtonMap[buttonConfig["focus"]]) {
		speed /= 2
	}
	var tx, ty float64

	if ebiten.IsKeyPressed(controlmap.KeyMap[keyConfig["right"]]) ||
		ebiten.IsGamepadButtonPressed(gamepadID, controlmap.ButtonMap[buttonConfig["right"]]) ||
		ebiten.GamepadAxis(gamepadID, 0) > deadZone {
		tx = 1
	}
	if ebiten.IsKeyPressed(controlmap.KeyMap[keyConfig["left"]]) ||
		ebiten.IsGamepadButtonPressed(gamepadID, controlmap.ButtonMap[buttonConfig["left"]]) ||
		ebiten.GamepadAxis(gamepadID, 0) < -deadZone {
		tx = -1
	}
	if ebiten.IsKeyPressed(controlmap.KeyMap[keyConfig["down"]]) ||
		ebiten.IsGamepadButtonPressed(gamepadID, controlmap.ButtonMap[buttonConfig["down"]]) ||
		ebiten.GamepadAxis(gamepadID, 1) > deadZone {
		ty = 1
	}
	if ebiten.IsKeyPressed(controlmap.KeyMap[keyConfig["up"]]) ||
		ebiten.IsGamepadButtonPressed(gamepadID, controlmap.ButtonMap[buttonConfig["up"]]) ||
		ebiten.GamepadAxis(gamepadID, 1) < -deadZone {
		ty = -1
	}

	if r := math.Sqrt(tx*tx + ty*ty); r != 0 {
		tx = tx / r * speed
		ty = ty / r * speed

		tx = math.Max(0, p.hitBox.x()+tx) - p.hitBox.x()
		tx = math.Min(windowWidth-float64(p.hitBoxSize), p.hitBox.x()+tx) - p.hitBox.x()

		ty = math.Max(0, p.hitBox.y()+ty) - p.hitBox.y()
		ty = math.Min(windowHeight-float64(p.hitBoxSize), p.hitBox.y()+ty) - p.hitBox.y()
	}

	p.skin.opts.GeoM.Translate(tx, ty)
	p.hitBox.opts.GeoM.Translate(tx, ty)

	p.vx = tx
	p.vy = ty

	return
}

func (p *player) shootBullets(freq int, n int, spreadDeg float64) {

	if frameCounter-p.lastShotFrame >= ebiten.MaxTPS()/p.bulletFreq {

		if indices := findNFirsts(p.bullets, n, func(b bullet) bool { return !b.isOnScreen }); len(indices) == n {

			for i := range indices {
				var angleDeg float64
				if n == 1 {
					angleDeg = 0
				} else {
					angleDeg = -spreadDeg/2 + float64(i)*spreadDeg/float64(n-1)
				}
				p.bullets[indices[i]].x = -float64(p.bulletSize)/2 + p.centreX() + p.bulletSpawnOffset*math.Sin(angleDeg*math.Pi/180)
				p.bullets[indices[i]].y = -float64(p.bulletSize)/2 + p.centreY() - p.bulletSpawnOffset*math.Cos(angleDeg*math.Pi/180)
				if p.isFocus {
					p.bullets[indices[i]].vx = 0
					p.bullets[indices[i]].vy = -1
				} else {
					p.bullets[indices[i]].vx = math.Sin(angleDeg * math.Pi / 180)
					p.bullets[indices[i]].vy = -math.Cos(angleDeg * math.Pi / 180)
				}
				p.bullets[indices[i]].xSize = p.bulletSize
				p.bullets[indices[i]].ySize = p.bulletSize
				p.bullets[indices[i]].isOnScreen = true
			}
		}
		p.lastShotFrame = frameCounter
	}

}

func initPlayer() player {
	var (
		skinSize          int     = 32
		hitBoxSize        int     = 8
		mvtSpeed          float64 = 8
		bulletSize        int     = 12
		bulletSpeed       float64 = 20
		bulletFreq        int     = 60
		baseBulletSpread  float64 = 100 //degrees
		focusBulletSpread float64 = 90  //degrees
		bulletStreams     int     = 9
		bulletSpawnOffset float64 = 60
	)

	skinImage, errS := ebiten.NewImage(skinSize, skinSize, ebiten.FilterNearest)
	utils.LogError(errS)
	skinImage.Fill(color.White)
	skinOpts := ebiten.DrawImageOptions{}
	skinOpts.GeoM.Translate(float64(hitBoxSize-skinSize)/2, float64(hitBoxSize-skinSize)/2)
	skinOpts.GeoM.Translate((windowWidth-float64(hitBoxSize))/2, (windowHeight-float64(hitBoxSize))/2)

	skin := sprite{
		skinImage,
		skinOpts,
	}

	hitBoxImage, errH := ebiten.NewImage(hitBoxSize, hitBoxSize, ebiten.FilterNearest)
	utils.LogError(errH)
	hitBoxImage.Fill(color.RGBA{255, 0, 0, 255})
	hitBoxOpts := ebiten.DrawImageOptions{}
	hitBoxOpts.GeoM.Translate((windowWidth-float64(hitBoxSize))/2, (windowHeight-float64(hitBoxSize))/2)

	hitBox := sprite{
		hitBoxImage,
		hitBoxOpts,
	}

	bullets := make([]bullet, maxBullets)

	bulletSkin, errB := ebiten.NewImage(bulletSize, bulletSize, ebiten.FilterNearest)
	utils.LogError(errB)
	bulletSkin.Fill(color.White)
	bulletSprite := sprite{bulletSkin, ebiten.DrawImageOptions{}}

	return player{
		skin,
		hitBox,
		bullets,
		bulletSkin,
		bulletSprite,
		0,     // lastShotFrame
		false, // isFocus

		0, 0, // Initialise at 0 speed

		skinSize,
		hitBoxSize,
		mvtSpeed,

		bulletSize,
		bulletSpeed,
		bulletFreq,
		baseBulletSpread,
		focusBulletSpread,
		bulletStreams,
		bulletSpawnOffset,
	}
}
