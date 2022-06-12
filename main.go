package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"runtime/pprof"
)

var log = false

type DNA struct {
	s     string
	len   int
	left  *DNA
	right *DNA
}

func dnaFromString(s string) *DNA {
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
	if d.left == nil {
		return d.s[i]
	} else {
		if i < d.left.len {
			return d.left.get(i)
		} else {
			return d.right.get(i - d.left.len)
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

func (d1 *DNA) append(d2 *DNA) *DNA {
	if d1.Len() == 0 {
		return d2
	}
	if d2.Len() == 0 {
		return d1
	}
	if d1.Len() < 100 && d2.left != nil && d2.left.left == nil && len(d2.left.s) < 100 {
		return &DNA{
			len:   d1.len + d2.len,
			left:  dnaFromString(d1.asString() + d2.left.s),
			right: d2.right,
		}
	}
	if d2.Len() < 100 && d1.right != nil && d1.right.right == nil && len(d1.right.s) < 100 {
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

func emptyDNA() *DNA {
	return &DNA{"", 0, nil, nil}
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
	cur := iter.stack[0]
	ret := cur.get(iter.i)
	iter.i++
	if iter.i == cur.Len() {
		iter.i = 0
		iter.stack = iter.stack[1:]
		if len(iter.stack) > 0 && iter.stack[0].left != nil {
			newstack := dnaToStack(iter.stack[0])
			iter.stack = append(newstack, iter.stack[1:]...)
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
	ret = ret.append(iter.stack[0].skip(iter.i))
	for i := 1; i < len(iter.stack); i++ {
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
	var stack []*DNA
	for d.right != nil {
		stack = append([]*DNA{d.right}, stack...)
		d = d.left
	}
	stack = append([]*DNA{d}, stack...)
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
		} else if iteration%10000 == 0 {
			fmt.Printf("iteration %d\n", iteration)
		} else if iteration > 10000000 {
			panic("done")
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
	var p Pattern
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
	var t Template
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
	curpart := ""
	for _, t := range tmpl {
		switch v := t.(type) {
		case int32:
			curpart += string(v)
		case int:
			x := 0
			if v < len(e) {
				x = e[v].Len()
			}
			curpart += asnat(x)
		case []int:
			if v[0] < len(e) {
				if len(curpart) > 0 {
					parts = append(parts, dnaFromString(curpart))
				}
				parts = append(parts, protect(v[1], e[v[0]]))
				curpart = ""
			}
		default:
			panic(fmt.Sprintf("unexpected template element: %v", reflect.TypeOf(t)))
		}
	}
	r := emptyDNA()
	for _, part := range parts {
		r = r.append(part)
	}
	if len(curpart) > 0 {
		r = r.append(dnaFromString(curpart))
	}
	dna = r.append(dna)
}

// 4 => IICP
// 0 => P
// 5 => CICP
func asnat(i int) string {
	ret := ""
	for {
		if i == 0 {
			ret += "P"
			return ret
		}
		if i%2 == 0 {
			ret += "I"
		} else {
			ret += "C"
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
	ret := ""
	for i := 0; i < d.Len(); i++ {
		switch d.get(i) {
		case 'I':
			ret += "C"
		case 'C':
			ret += "F"
		case 'F':
			ret += "P"
		case 'P':
			ret += "IC"
		default:
			panic("invalid base")
		}
	}
	return dnaFromString(ret)
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
}
