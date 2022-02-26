package main

import (
	"encoding/binary"
	"fmt"
	"github.com/alttpo/snes/mapping/lorom"
	"io/ioutil"
)

type Sprite struct {
	Kind    byte
	X       byte
	Y       byte
	SubKind byte
	Layer   byte

	HasDropItem bool
	DropsId     byte
}

func main() {
	var err error

	fileName := "alttp.smc"

	var contents []byte
	contents, err = ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	vanilla_sprites := make(map[uint16][]Sprite)

	le := binary.LittleEndian
	var offs uint32

	const room_sprite_pointer = 0x04_C298
	spritePointer := 0x04_0000 + uint32(le.Uint16(contents[room_sprite_pointer:room_sprite_pointer+2]))
	for supertile := uint16(0); supertile < 0x0128; supertile++ {
		offs = spritePointer + (uint32(supertile) * 2)
		spriteAddressSnes := 0x09_0000 + uint32(le.Uint16(contents[offs:offs+2]))

		var spriteAddressPC uint32
		spriteAddressPC, err = lorom.BusAddressToPak(spriteAddressSnes)
		spriteAddressPC++

		sprites := make([]Sprite, 0, 10)
		for i := 0; ; i++ {
			b1 := contents[spriteAddressPC+0]
			b2 := contents[spriteAddressPC+1]
			b3 := contents[spriteAddressPC+2]
			spriteAddressPC += 3

			if b1 == 0xFF {
				break
			}

			spr := Sprite{
				Kind:    b3,
				X:       b2 & 0x1F,
				Y:       b1 & 0x1F,
				SubKind: ((b2 & 0xE0) >> 5) + ((b1 & 0x60) >> 2),
				Layer:   (b1 & 0x80) >> 7,
				DropsId: 0,
			}

			//if (spr.id == 0xE4 && spr.x == 0x00 && spr.y == 0x1E && spr.layer == 1 && ((spr.subtype)) == 0x18)

			// does subtype=0x18 indicate item drop?
			if spr.SubKind == 0x18 && spr.Layer == 1 {
				sprites[i-1].HasDropItem = true
				sprites[i-1].DropsId = spr.Kind
				if spr.Y == 0x1D {
					// big key?
					sprites[i-1].DropsId = 0xE5
				}
			} else {
				sprites = append(sprites, spr)
			}
		}

		vanilla_sprites[supertile] = sprites

		if false {
			fmt.Printf("\nsupertile 0x%04x:\n", supertile)
			for i, spr := range sprites {
				drops := ""
				if spr.HasDropItem {
					drops = fmt.Sprintf(" drops %02x", spr.DropsId)
				}
				fmt.Printf("  [%2d]: %02x (%02x) at %02x,%02x%s\n", i, spr.Kind, spr.SubKind, spr.X, spr.Y, drops)
			}
		}
	}

	fmt.Println(`
class Sprite(object):
	def __init__(self, super_tile, kind, sub_type, layer, tile_x, tile_y, drops_item=False, drop_item_kind=None):
		self.super_tile = super_tile
		self.kind = kind
		self.sub_type = sub_type
		self.layer = layer
		self.tile_x = tile_x
		self.tile_y = tile_y
		self.drops_item = drops_item
		self.drop_item_kind = drop_item_kind

# map of super_tile to list of Sprite objects:
vanilla_sprites = {}

def create_sprite(super_tile, kind, sub_type, layer, tile_x, tile_y, drops_item=False, drop_item_kind=None):
	if super_tile not in vanilla_sprites:
		vanilla_sprites[super_tile] = []
	vanilla_sprites[super_tile].append(Sprite(kind, sub_type, layer, tile_x, tile_y, drops_item, drop_item_kind))
`)

	for supertile := uint16(0); supertile < 0x0128; supertile++ {
		sprites := vanilla_sprites[supertile]
		for _, spr := range sprites {
			drops := ""
			if spr.HasDropItem {
				drops = fmt.Sprintf(", True, 0x%02x", spr.DropsId)
			}
			fmt.Printf("create_sprite(0x%04x, 0x%02x, 0x%02x, %d, 0x%02x, 0x%02x%s)\n", supertile, spr.Kind, spr.SubKind, spr.Layer, spr.X, spr.Y, drops)
		}
	}
}
