package main

import (
	"encoding/binary"
	"fmt"
	"github.com/alttpo/snes/mapping/lorom"
	"io/ioutil"
)

func main() {
	var err error

	fileName := "alttp.smc"

	var contents []byte
	contents, err = ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	le := binary.LittleEndian
	var offs uint32

	const room_sprite_pointer = 0x04_C298
	spritePointer := 0x04_0000 + uint32(le.Uint16(contents[room_sprite_pointer:room_sprite_pointer+2]))

	for supertile := uint16(0); supertile < 0x0128; supertile++ {
		fmt.Printf("\nsupertile 0x%04x:\n", supertile)

		offs = spritePointer + (uint32(supertile) * 2)
		spriteAddressSnes := 0x09_0000 + uint32(le.Uint16(contents[offs:offs+2]))

		var spriteAddressPC uint32
		spriteAddressPC, err = lorom.BusAddressToPak(spriteAddressSnes)
		spriteAddressPC++

		type Sprite struct {
			Id      byte
			X       byte
			Y       byte
			Subtype byte
			Layer   byte

			HasDropItem bool
			DropsId     byte
		}

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
				Id:      b3,
				X:       b2 & 0x1F,
				Y:       b1 & 0x1F,
				Subtype: ((b2 & 0xE0) >> 5) + ((b1 & 0x60) >> 2),
				Layer:   (b1 & 0x80) >> 7,
				DropsId: 0,
			}

			//if (spr.id == 0xE4 && spr.x == 0x00 && spr.y == 0x1E && spr.layer == 1 && ((spr.subtype)) == 0x18)

			// does subtype=0x18 indicate item drop?
			if spr.Subtype == 0x18 && spr.Layer == 1 {
				sprites[i-1].HasDropItem = true
				sprites[i-1].DropsId = spr.Id
				if spr.Y == 0x1D {
					// big key?
					sprites[i-1].DropsId = 0xE5
				}
			} else {
				sprites = append(sprites, spr)
			}
		}

		for i, spr := range sprites {
			drops := ""
			if spr.HasDropItem {
				drops = fmt.Sprintf(" drops %02x", spr.DropsId)
			}
			fmt.Printf("  [%2d]: %02x (%02x) at %02x,%02x%s\n", i, spr.Id, spr.Subtype, spr.X, spr.Y, drops)
		}
	}
}
