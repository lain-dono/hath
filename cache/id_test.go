package cache

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
	"time"
)

func Test(t *testing.T) {
	jpg := Id{"02c828887df7a746371c94c03b952d01a4ee041e", 555101, 995, 1400, "jpg"}
	png := Id{"028375865bd89c5f98107bcb974446364d1586e1", 566576, 800, 600, "png"}
	gif := Id{"f264e78cb64ba012a70e479c1edbd33c4a58aa64", 9743143, 800, 600, "gif"}

	jpg_t, _ := time.Parse(time.RFC822, "01 Jan 06 15:04 MSK")
	png_t, _ := time.Parse(time.RFC822, "02 Jan 08 10:04 MSK")
	gif_t, _ := time.Parse(time.RFC822, "03 Jan 08 17:04 MSK")

	checkId := func(other Id, mime string) func() {
		return func() {
			id, ok := NewIdFromString(other.String())
			So(ok, ShouldBeTrue)
			So(id, ShouldResemble, other)
			So(id.String(), ShouldEqual, other.String())
			So(id.MimeType(), ShouldEqual, mime)
			So(id.Hash(), ShouldEqual, id.Hash())
			So(id.Size(), ShouldEqual, id.Size())
			x, y := id.Res()
			ox, oy := other.Res()
			So(x, ShouldEqual, ox)
			So(y, ShouldEqual, oy)
		}
	}

	Convey("All", t, func() {
		Convey("Id", func() {
			Convey("jpg", checkId(jpg, CONTENT_TYPE_JPG))
			Convey("png", checkId(png, CONTENT_TYPE_PNG))
			Convey("gif", checkId(gif, CONTENT_TYPE_GIF))
		})
		Convey("DB", func() {
			var lasthit time.Time
			db := NewDB("data.db")
			defer os.Remove("data.db")
			db.Optimize()

			db.InsertCachedFile(jpg, jpg_t)
			db.InsertCachedFile(png, png_t)
			db.InsertCachedFile(gif, gif_t)

			Convey("LastHit", func() {
				lasthit, _ = db.LastHit(jpg)
				So(lasthit.Equal(jpg_t), ShouldBeTrue)
				lasthit, _ = db.LastHit(png)
				So(lasthit.Equal(png_t), ShouldBeTrue)
				lasthit, _ = db.LastHit(gif)
				So(lasthit.Equal(gif_t), ShouldBeTrue)

				Convey("SetLastHit", func() {
					t2, _ := time.Parse(time.RFC822, "02 Jan 06 10:14 MSK")
					db.SetLastHit(jpg, t2)
					lasthit, _ = db.LastHit(jpg)
					So(lasthit.Equal(t2), ShouldBeTrue)

					db.SetLastHit(jpg, jpg_t)
					lasthit, _ = db.LastHit(jpg)
					So(lasthit.Equal(jpg_t), ShouldBeTrue)
				})
			})
			Convey("CountStats", func() {
				count, size, _ := db.CountStats()
				So(count, ShouldEqual, 3)
				So(size, ShouldEqual, jpg.Size()+png.Size()+gif.Size())
			})
			Convey("Active & Remove", func() {
				// test active by default
				db.Activate(jpg)
				db.Activate(gif)
				n, _ := db.RemoveInactive()
				So(n, ShouldEqual, 0)

				db.ClearActive()
				db.Activate(jpg)
				db.Activate(gif)
				n, _ = db.RemoveInactive()
				So(n, ShouldEqual, 1)
				hist, _ := db.CachedFileSortOnLasthit(0, 5)
				So(hist[0].Id.Hash(), ShouldEqual, jpg.Hash())
				So(hist[1].Id.Hash(), ShouldEqual, gif.Hash())
			})

			/*
				file, err := db.CachedFileByLastHit()
				Println(err)
			*/
		})
	})
}
