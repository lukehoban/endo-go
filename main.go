package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"runtime/pprof"
	"strings"
)

const log = false

type DNA struct {
	s     string
	len   int
	left  *DNA
	right *DNA
}

func dnaFromString(s string) *DNA {
	if s == "" {
		return emptyDNA()
	}
	if len(s) < 4 {
		return dnaCache[s]
	}
	return &DNA{
		s:     s,
		len:   len(s),
		left:  nil,
		right: nil,
	}
}

func (d *DNA) get(i int) byte {
	if i >= d.len {
		return 0
	}
	for {
		if d.left == nil {
			return d.s[i]
		}
		if i < d.left.len {
			d = d.left
		} else {
			i -= d.left.len
			d = d.right
		}
	}
}

func (d *DNA) skip(i int) *DNA {
	if i == 0 {
		return d
	}
	if d.left == nil {
		return dnaFromString(d.s[i:])
	} else {
		if i < d.left.len {
			return &DNA{
				s:     "",
				len:   d.len - i,
				left:  d.left.skip(i),
				right: d.right,
			}
		} else {
			return d.right.skip(i - d.left.len)
		}
	}
}

const SMALL = 50

func (d1 *DNA) append(d2 *DNA) *DNA {
	if d1.Len() == 0 {
		return d2
	}
	if d2.Len() == 0 {
		return d1
	}
	if d1.Len() < SMALL && d2.left != nil && d2.left.left == nil && len(d2.left.s) < SMALL {
		return &DNA{
			len:   d1.len + d2.len,
			left:  dnaFromString(d1.asString() + d2.left.s),
			right: d2.right,
		}
	}
	if d2.Len() < SMALL && d1.right != nil && d1.right.right == nil && len(d1.right.s) < SMALL {
		return &DNA{
			len:   d1.len + d2.len,
			left:  d1.left,
			right: dnaFromString(d1.right.s + d2.asString()),
		}
	}
	return &DNA{
		s:     "",
		len:   d1.len + d2.len,
		left:  d1,
		right: d2,
	}
}

func (d *DNA) keep(n int) *DNA {
	if n == d.Len() {
		return d
	}
	if d.left == nil {
		return dnaFromString(d.s[0:n])
	} else {
		if n > d.left.len {
			return &DNA{
				s:     "",
				len:   n,
				left:  d.left,
				right: d.right.keep(n - d.left.len),
			}
		} else {
			return d.left.keep(n)
		}
	}
}

var empty = DNA{}

func emptyDNA() *DNA {
	return &empty
}

func (d *DNA) Len() int {
	return d.len
}

func (d *DNA) substring(i int, end int) *DNA {
	if end > d.Len() {
		end = d.Len()
	}
	return d.skip(i).keep(end - i)
}

func (d *DNA) asString() string {
	if d.left == nil {
		return d.s
	} else {
		return d.left.asString() + d.right.asString()
	}
}

func (d *DNA) String() string {
	snip := d.substring(0, 10)
	cont := ""
	if d.Len() > 10 {
		cont = "..."
	}
	return fmt.Sprintf("%s%s (%d bases)", snip.asString(), cont, d.Len())
}

type DNAIterator struct {
	stack []*DNA
	i     int
	buf   []byte
}

func (iter *DNAIterator) Next() byte {
	if len(iter.buf) > 0 {
		ret := iter.buf[0]
		iter.buf = iter.buf[1:]
		return ret
	}
	return iter.next()
}

func (iter *DNAIterator) next() byte {
	if len(iter.stack) == 0 {
		return 0
	}
	lastStackIndex := len(iter.stack) - 1
	cur := iter.stack[lastStackIndex]
	if iter.i >= len(cur.s) {
		return 0
	}
	ret := cur.s[iter.i]
	iter.i++
	if iter.i == cur.Len() {
		iter.i = 0
		iter.stack = iter.stack[:lastStackIndex]
		lastStackIndex--
		if len(iter.stack) > 0 && iter.stack[lastStackIndex].left != nil {
			newstack := dnaToStack(iter.stack[lastStackIndex])
			iter.stack = append(iter.stack[:lastStackIndex], newstack...)
		}
	}
	return ret
}

func (iter *DNAIterator) Peek() byte {
	b := iter.next()
	iter.buf = append(iter.buf, b)
	return b
}

func (iter *DNAIterator) Rest() *DNA {
	ret := dnaFromString(string(iter.buf))
	if len(iter.stack) == 0 {
		return ret
	}
	ret = ret.append(iter.stack[len(iter.stack)-1].skip(iter.i))
	for i := len(iter.stack) - 2; i >= 0; i-- {
		ret = &DNA{
			s:     "",
			len:   ret.len + iter.stack[i].len,
			left:  ret,
			right: iter.stack[i],
		}
	}
	return ret
}

func dnaToStack(d *DNA) []*DNA {
	// var stack []*DNA
	stack := make([]*DNA, 0, 10)
	for d.right != nil {
		stack = append(stack, d.right)
		d = d.left
	}
	stack = append(stack, d)
	return stack
}

func (d *DNA) iterator() *DNAIterator {
	return &DNAIterator{
		stack: dnaToStack(d),
		i:     0,
	}
}

var dna *DNA
var rna []string

type Pattern []interface{}
type Template []interface{}

func (pat Pattern) String() string {
	ret := ""
	for _, p := range pat {
		switch v := p.(type) {
		case int32:
			ret += string(v)
		case int:
			ret += fmt.Sprintf("!%d", v)
		case bool:
			if v {
				ret += "("
			} else {
				ret += ")"
			}
		case string:
			ret += fmt.Sprintf("?%q", v)
		default:
			panic(fmt.Sprintf("unexpected pattern element: %v", reflect.TypeOf(p)))
		}
	}
	return ret
}

func (tmpl Template) String() string {
	ret := ""
	for _, t := range tmpl {
		switch v := t.(type) {
		case int32:
			ret += string(v)
		case int:
			ret += fmt.Sprintf("|%d|", v)
		case []int:
			if v[1] == 0 {
				ret += fmt.Sprintf("\\%d", v[0])
			} else {
				ret += fmt.Sprintf("\\%d(%d)", v[0], v[1])
			}
		default:
			panic(fmt.Sprintf("unexpected template element: %v", reflect.TypeOf(t)))
		}
	}
	return ret
}

func do(prefix string) error {
	byts, err := ioutil.ReadFile("./endo.dna")
	if err != nil {
		return err
	}

	dna = dnaFromString(prefix).append(dnaFromString(string(byts)))

	iteration := 0
	for {

		if log {
			fmt.Printf("\niteration %d\n", iteration)
			fmt.Printf("dna = %s\n", dna)
		} else if iteration%100000 == 0 {
			fmt.Printf("iteration %d\n", iteration)
		}

		iter := dna.iterator()

		pat, err := pattern(iter)
		if err != nil {
			return err
		}
		if log {
			fmt.Printf("pat = %s\n", pat)
		}

		tmpl, err := template(iter)
		if err != nil {
			return err
		}
		if log {
			fmt.Printf("tmpl = %s\n", tmpl)
		}

		dna = iter.Rest()

		matchreplace(pat, tmpl)
		iteration++
	}
}

func pattern(iter *DNAIterator) (Pattern, error) {
	p := pat[:0]
	lvl := 0
	for {
		switch iter.Next() {
		case 'C':
			p = append(p, 'I')
		case 'F':
			p = append(p, 'C')
		case 'P':
			p = append(p, 'F')
		case 'I':
			switch iter.Next() {
			case 'C':
				p = append(p, 'P')
			case 'P':
				n, err := nat(iter)
				if err != nil {
					return p, err
				}
				p = append(p, n)
			case 'F':
				iter.Next()
				c := consts(iter)
				p = append(p, c)
			case 'I':
				switch iter.Next() {
				case 'P':
					lvl++
					p = append(p, true)
				case 'C', 'F':
					if lvl == 0 {
						return p, nil
					}
					lvl--
					p = append(p, false)
				case 'I':
					r := ""
					for i := 0; i < 7; i++ {
						r += fmt.Sprintf("%c", iter.Next())
					}
					rna = append(rna, r)
				default:
					return p, fmt.Errorf("end of file pat 1")
				}
			default:
				return p, fmt.Errorf("end of file pat 2")
			}
		default:
			return p, fmt.Errorf("end of file pat 3")
		}
	}
}

// CICP => 5
// ICP =>  2
// CIP =>  1
func nat(iter *DNAIterator) (int, error) {
	ret := 0
	i := 0
	for {
		switch iter.Next() {
		case 'C':
			ret += 1 << i
		case 'I', 'F':
		case 'P':
			return ret, nil
		default:
			return 0, fmt.Errorf("end of file nat")
		}
		i++
	}
}

// CFICP => ICPF
func consts(iter *DNAIterator) string {
	str := ""
	for {
		switch iter.Peek() {
		case 'C':
			str += "I"
		case 'F':
			str += "C"
		case 'P':
			str += "F"
		case 'I':
			if iter.Peek() != 'C' {
				return str
			}
			str += "P"
			iter.Next()
		}
		iter.Next()
	}
}

func template(iter *DNAIterator) (Template, error) {
	t := tmpl[:0]
	for {
		switch iter.Next() {
		case 'C':
			t = append(t, 'I')
		case 'F':
			t = append(t, 'C')
		case 'P':
			t = append(t, 'F')
		case 'I':
			switch iter.Next() {
			case 'C':
				t = append(t, 'P')
			case 'F', 'P':
				l, err := nat(iter)
				if err != nil {
					return t, err
				}
				n, err := nat(iter)
				if err != nil {
					return t, err
				}
				t = append(t, []int{n, l})
			case 'I':
				switch iter.Next() {
				case 'C', 'F':
					return t, nil
				case 'P':
					n, err := nat(iter)
					if err != nil {
						return t, err
					}
					t = append(t, n)
				case 'I':
					r := ""
					for i := 0; i < 7; i++ {
						r += fmt.Sprintf("%c", iter.Next())
					}
					rna = append(rna, r)
				default:
					return t, fmt.Errorf("end of file")
				}
			default:
				return t, fmt.Errorf("end of file")
			}
		default:
			return t, fmt.Errorf("end of file")
		}
	}
}

func matchreplace(pat Pattern, tmpl Template) {
	i := 0
	var e []*DNA
	var c []int

	for _, p := range pat {
		switch v := p.(type) {
		case int32:
			if dna.get(i) == byte(v) {
				i++
			} else {
				return
			}
		case int:
			i += v
			if i > dna.Len() {
				return
			}
		case string:
			n := findPostfix(dna.skip(i), v)
			if n == -1 {
				return
			}
			i += n
		case bool:
			if v {
				c = append([]int{i}, c...)
			} else {
				e = append(e, dna.substring(c[0], i))
				c = c[1:]
			}
		default:
			panic("nyi - matchreplace - hmm")
		}
	}

	dna = dna.skip(i)
	replace(tmpl, e)
}

// DNA = ICFICFICPF
// needle = ICP
// ret = 8
func findPostfix(d *DNA, needle string) int {
outer:
	for i := 0; i <= d.Len()-len(needle); i++ {
		for j := 0; j < len(needle); j++ {
			if d.get(i+j) != needle[j] {
				continue outer
			}
		}
		return i + len(needle)
	}
	return -1
}

func replace(tmpl Template, e []*DNA) {
	var parts []*DNA
	var curpart strings.Builder
	for _, t := range tmpl {
		switch v := t.(type) {
		case int32:
			curpart.WriteByte(byte(v))
		case int:
			x := 0
			if v < len(e) {
				x = e[v].Len()
			}
			curpart.Write([]byte(asnat(x)))
		case []int:
			if v[0] < len(e) {
				if curpart.Len() > 0 {
					parts = append(parts, dnaFromString(curpart.String()))
				}
				parts = append(parts, protect(v[1], e[v[0]]))
				curpart = strings.Builder{}
			}
		default:
			panic(fmt.Sprintf("unexpected template element: %v", reflect.TypeOf(t)))
		}
	}
	r := emptyDNA()
	for _, part := range parts {
		r = r.append(part)
	}
	if curpart.Len() > 0 {
		r = r.append(dnaFromString(curpart.String()))
	}
	dna = r.append(dna)
}

// 4 => IICP
// 0 => P
// 5 => CICP
func asnat(i int) string {
	var ret strings.Builder
	for {
		if i == 0 {
			ret.WriteByte('P')
			return ret.String()
		}
		if i%2 == 0 {
			ret.WriteByte('I')
		} else {
			ret.WriteByte('C')
		}
		i /= 2
	}
}

func protect(l int, d *DNA) *DNA {
	for i := 0; i < l; i++ {
		d = quote(d)
	}
	return d
}

// ICFPI => CFPICC
func quote(d *DNA) *DNA {
	var ret strings.Builder
	for i := 0; i < d.Len(); i++ {
		switch d.get(i) {
		case 'I':
			ret.WriteByte('C')
		case 'C':
			ret.WriteByte('F')
		case 'F':
			ret.WriteByte('P')
		case 'P':
			ret.WriteByte('I')
			ret.WriteByte('C')
		default:
			panic("invalid base")
		}
	}
	return dnaFromString(ret.String())
}

var pat Pattern
var tmpl Template
var dnaCache map[string]*DNA = map[string]*DNA{}

func init() {
	pat = make(Pattern, 0, 1024)
	tmpl = make(Template, 0, 1024)
	chars := []byte{'I', 'C', 'F', 'P'}
	for _, x0 := range chars {
		s := string([]byte{x0})
		dnaCache[s] = &DNA{
			s:   s,
			len: 1,
		}
		for _, x1 := range chars {
			s := string([]byte{x0, x1})
			dnaCache[s] = &DNA{
				s:   s,
				len: 2,
			}
			for _, x2 := range chars {
				s := string([]byte{x0, x1, x2})
				dnaCache[s] = &DNA{
					s:   s,
					len: 3,
				}
			}
		}
	}
}

type Coord int
type Pos struct {
	X, Y Coord
}
type Component int
type RGB struct {
	R, G, B Component
}
type Transparency Component
type Pixel struct {
	RGB RGB
	T   Transparency
}
type Bitmap [600][600]Pixel
type Bucket []interface{}
type Dir string

const (
	N Dir = "N"
	E Dir = "E"
	S Dir = "S"
	W Dir = "W"
)

var (
	Black       RGB          = RGB{0, 0, 0}
	Red         RGB          = RGB{255, 0, 0}
	Green       RGB          = RGB{0, 255, 0}
	Yellow      RGB          = RGB{255, 255, 0}
	Blue        RGB          = RGB{0, 0, 255}
	Magenta     RGB          = RGB{255, 0, 255}
	Cyan        RGB          = RGB{0, 255, 255}
	White       RGB          = RGB{255, 255, 255}
	Transparent Transparency = 0
	Opaque      Transparency = 255
)

func build() (Bitmap, error) {
	var bucket Bucket
	dir := E
	pos := Pos{0, 0}
	mark := Pos{0, 0}
	bitmaps := []Bitmap{Bitmap{}}

	currentPixel := func() Pixel {
		var numrgb, rc, gc, bc Component
		var numt, ac Transparency
		for i := range bucket {
			switch v := bucket[i].(type) {
			case RGB:
				numrgb++
				rc += v.R
				gc += v.G
				bc += v.B
			case Transparency:
				numt++
				ac += v
			default:
				panic("invalid type in bucket")
			}
		}
		if numrgb > 0 {
			rc /= numrgb
			gc /= numrgb
			bc /= numrgb
		}
		if numt > 0 {
			ac /= numt
		} else {
			ac = 255
		}
		return Pixel{
			RGB: RGB{(rc * Component(ac)) / 255, (gc * Component(ac)) / 255, (bc * Component(ac)) / 255},
			T:   ac,
		}
	}

	getPixel := func(p Pos) Pixel {
		return bitmaps[0][p.X][p.Y]
	}

	setPixel := func(p Pos) {
		bitmaps[0][p.X][p.Y] = currentPixel()
	}

	for _, r := range rna {
		switch r {
		case "PIPIIIC":
			bucket = append(bucket, Black)
		case "PIPIIIP":
			bucket = append(bucket, Red)
		case "PIPIICC":
			bucket = append(bucket, Green)
		case "PIPIICF":
			bucket = append(bucket, Yellow)
		case "PIPIICP":
			bucket = append(bucket, Blue)
		case "PIPIIFC":
			bucket = append(bucket, Magenta)
		case "PIPIIFF":
			bucket = append(bucket, Cyan)
		case "PIPIIPC":
			bucket = append(bucket, White)
		case "PIPIIPF":
			bucket = append(bucket, Transparent)
		case "PIPIIPP":
			bucket = append(bucket, Opaque)
		case "PIIPICP":
			bucket = nil
		case "PIIIIIP":
			switch dir {
			case N:
				pos.Y = (pos.Y + 599) % 600
			case E:
				pos.X = (pos.X + 1) % 600
			case S:
				pos.Y = (pos.Y + 1) % 600
			case W:
				pos.X = (pos.X + 599) % 600
			}
		case "PCCCCCP":
			switch dir {
			case N:
				dir = W
			case E:
				dir = N
			case S:
				dir = E
			case W:
				dir = S
			}
		case "PFFFFFP":
			switch dir {
			case N:
				dir = E
			case E:
				dir = S
			case S:
				dir = W
			case W:
				dir = N
			}
		case "PCCIFFP":
			mark = pos
		case "PFFICCP":
			p0 := pos
			p1 := mark
			deltax := p1.X - p0.X
			deltay := p1.Y - p0.Y
			d := deltax
			if -deltax > d {
				d = -deltax
			}
			if deltay > d {
				d = deltay
			}
			if -deltay > d {
				d = -deltay
			}
			var c Coord = 0
			if deltax*deltay <= 0 {
				c = 1
			}
			x := p0.X*d + (d-c)/2
			y := p0.Y*d + (d-c)/2
			for i := Coord(0); i < d; i++ {
				setPixel(Pos{x / d, y / d})
				x += deltax
				y += deltay
			}
			setPixel(p1)
		case "PIIPIIP":
			fill := func(p Pos, initial Pixel) {
				toFill := []Pos{p}
				for len(toFill) > 0 {
					p = toFill[0]
					toFill = toFill[1:]
					if getPixel(p) == initial {
						setPixel(p)
						if p.X > 0 {
							toFill = append(toFill, Pos{p.X - 1, p.Y})
						}
						if p.X < 599 {
							toFill = append(toFill, Pos{p.X + 1, p.Y})
						}
						if p.Y > 0 {
							toFill = append(toFill, Pos{p.X, p.Y - 1})
						}
						if p.Y < 599 {
							toFill = append(toFill, Pos{p.X, p.Y + 1})
						}
					}
				}
			}
			old := getPixel(pos)
			if currentPixel() != old {
				fill(pos, old)
			}
		case "PCCPFFP":
			if len(bitmaps) < 10 {
				bitmaps = append([]Bitmap{Bitmap{}}, bitmaps...)
			}
		case "PFFPCCP":
			if len(bitmaps) >= 2 {
				for y := Coord(0); y < 600; y++ {
					for x := Coord(0); x < 600; x++ {
						p0 := bitmaps[0][x][y]
						p1 := bitmaps[1][x][y]
						bitmaps[1][x][y] = Pixel{
							RGB: RGB{
								p0.RGB.R + ((p1.RGB.R * (255 - Component(p0.T))) / 255),
								p0.RGB.G + ((p1.RGB.G * (255 - Component(p0.T))) / 255),
								p0.RGB.B + ((p1.RGB.B * (255 - Component(p0.T))) / 255),
							},
							T: p0.T + ((p1.T * (255 - p0.T)) / 255),
						}
					}
				}
				bitmaps = bitmaps[1:]
			}
		case "PFFICCF":
			return bitmaps[0], fmt.Errorf("clip ()")
		default:
			// Do nothing
		}
	}
	return bitmaps[0], nil
}

func main() {

	if true {
		f, err := os.Create("out.prof")
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	prefix := ""
	if len(os.Args) >= 2 {
		prefix = os.Args[1]
	}
	err := do(prefix)
	if err != nil {
		fmt.Printf("#rna = %d\n", len(rna))
		fmt.Printf("%v\n", err)
	}

	bitmap, err := build()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	var builder strings.Builder
	for j := 0; j < 600; j++ {
		for i := 0; i < 600; i++ {
			pixel := bitmap[i][j]
			if pixel.RGB.R == 0 && pixel.RGB.G == 0 && pixel.RGB.B == 0 {
				builder.WriteByte(' ')
			} else {
				if pixel.RGB.R > pixel.RGB.G && pixel.RGB.R > pixel.RGB.B {
					builder.WriteByte('*')
				} else if pixel.RGB.G > pixel.RGB.B {
					builder.WriteByte('.')
				} else {
					builder.WriteByte('#')
				}
			}
		}
		builder.WriteByte('\n')
	}
	fmt.Printf("%s", builder.String())

}
