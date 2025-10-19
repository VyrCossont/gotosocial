// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package media

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"path"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/image/webp"
)

func BenchmarkFlipH(b *testing.B) {
	benchmarkFunc(b, func(img image.Image) {
		_ = flipH(img)
	})
}

func BenchmarkFlipV(b *testing.B) {
	benchmarkFunc(b, func(img image.Image) {
		_ = flipV(img)
	})
}

func BenchmarkRotate90(b *testing.B) {
	benchmarkFunc(b, func(img image.Image) {
		_ = rotate90(img)
	})
}

func BenchmarkRotate180(b *testing.B) {
	benchmarkFunc(b, func(img image.Image) {
		_ = rotate180(img)
	})
}

func BenchmarkRotate270(b *testing.B) {
	benchmarkFunc(b, func(img image.Image) {
		_ = rotate270(img)
	})
}

func BenchmarkTranspose(b *testing.B) {
	benchmarkFunc(b, func(img image.Image) {
		_ = transpose(img)
	})
}

func BenchmarkTransverse(b *testing.B) {
	benchmarkFunc(b, func(img image.Image) {
		_ = transverse(img)
	})
}

func BenchmarkResizeHorizontalLinear(b *testing.B) {
	benchmarkFunc(b, func(img image.Image) {
		_ = resizeHorizontalLinear(img, 64)
	})
}

func BenchmarkResizeVerticalLinear(b *testing.B) {
	benchmarkFunc(b, func(img image.Image) {
		_ = resizeVerticalLinear(img, 64)
	})
}

func benchmarkFunc(b *testing.B, fn func(image.Image)) {
	b.Helper()
	for _, testcase := range []struct {
		Path   string
		Decode func(io.Reader) (image.Image, error)
	}{
		{
			Path:   "./test/big-panda.gif",
			Decode: gif.Decode,
		},
		{
			Path:   "./test/clock-original.gif",
			Decode: gif.Decode,
		},
		{
			Path:   "./test/test-jpeg.jpg",
			Decode: jpeg.Decode,
		},
		{
			Path:   "./test/test-png-noalphachannel.png",
			Decode: png.Decode,
		},
		{
			Path:   "./test/test-png-alphachannel.png",
			Decode: png.Decode,
		},
		{
			Path:   "./test/rainbow-original.png",
			Decode: png.Decode,
		},
		{
			Path:   "./test/nb-flag-original.webp",
			Decode: webp.Decode,
		},
	} {
		file, err := openRead(testcase.Path)
		if err != nil {
			panic(err)
		}

		img, err := testcase.Decode(file)
		if err != nil {
			panic(err)
		}

		info, err := file.Stat()
		if err != nil {
			panic(err)
		}

		file.Close()

		testname := fmt.Sprintf("ext=%s type=%s size=%d",
			strings.TrimPrefix(path.Ext(testcase.Path), "."),
			strings.TrimPrefix(reflect.TypeOf(img).String(), "*image."),
			info.Size(),
		)

		b.Run(testname, func(b *testing.B) {
			b.Helper()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					fn(img)
				}
			})
		})
	}
}
