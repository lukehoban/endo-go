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

var dna *DNA
var rna []*DNA

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

		pat, err := pattern()
		if err != nil {
			return err
		}
		if log {
			fmt.Printf("pat = %s\n", pat)
		}

		tmpl, err := template()
		if err != nil {
			return err
		}
		if log {
			fmt.Printf("tmpl = %s\n", tmpl)
		}

		matchreplace(pat, tmpl)
		iteration++
	}
}

func pattern() (Pattern, error) {
	var p Pattern
	lvl := 0

	for {
		switch dna.get(0) {
		case 'C':
			dna = dna.skip(1)
			p = append(p, 'I')
		case 'F':
			dna = dna.skip(1)
			p = append(p, 'C')
		case 'P':
			dna = dna.skip(1)
			p = append(p, 'F')
		case 'I':
			switch dna.get(1) {
			case 'C':
				dna = dna.skip(2)
				p = append(p, 'P')
			case 'P':
				dna = dna.skip(2)
				n, err := nat()
				if err != nil {
					return p, err
				}
				p = append(p, n)
			case 'F':
				dna = dna.skip(3)
				c := consts()
				p = append(p, c)
			case 'I':
				switch dna.get(2) {
				case 'P':
					dna = dna.skip(3)
					lvl++
					p = append(p, true)
				case 'C', 'F':
					dna = dna.skip(3)
					if lvl == 0 {
						return p, nil
					}
					lvl--
					p = append(p, false)
				case 'I':
					rna = append(rna, dna.substring(3, 10))
					dna = dna.skip(10)
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
func nat() (int, error) {
	ret := 0
	i := 0
	for {
		switch dna.get(i) {
		case 'C':
			ret += 1 << i
		case 'I', 'F':
		case 'P':
			dna = dna.skip(i + 1)
			return ret, nil
		default:
			return 0, fmt.Errorf("end of file nat")
		}
		i++
	}
}

// CFICP => ICPF
func consts() string {
	str := ""
	i := 0
	for {
		switch dna.get(i) {
		case 'C':
			str += "I"
		case 'F':
			str += "C"
		case 'P':
			str += "F"
		case 'I':
			i++
			if dna.get(i) != 'C' {
				dna = dna.skip(i - 1)
				return str
			}
			str += "P"
		}
		i++
	}
}

func template() (Template, error) {
	var t Template
	for {
		switch dna.get(0) {
		case 'C':
			dna = dna.skip(1)
			t = append(t, 'I')
		case 'F':
			dna = dna.skip(1)
			t = append(t, 'C')
		case 'P':
			dna = dna.skip(1)
			t = append(t, 'F')
		case 'I':
			switch dna.get(1) {
			case 'C':
				dna = dna.skip(2)
				t = append(t, 'P')
			case 'F', 'P':
				dna = dna.skip(2)
				l, err := nat()
				if err != nil {
					return t, err
				}
				n, err := nat()
				if err != nil {
					return t, err
				}
				t = append(t, []int{n, l})
			case 'I':
				switch dna.get(2) {
				case 'C', 'F':
					dna = dna.skip(3)
					return t, nil
				case 'P':
					dna = dna.skip(3)
					n, err := nat()
					if err != nil {
						return t, err
					}
					t = append(t, n)
				case 'I':
					rna = append(rna, dna.substring(3, 10))
					dna = dna.skip(10)
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
	r := emptyDNA()
	for _, t := range tmpl {
		switch v := t.(type) {
		case int32:
			r = r.append(dnaFromString(string(v)))
		case int:
			x := 0
			if v < len(e) {
				x = e[v].Len()
			}
			r = r.append(dnaFromString(asnat(x)))
		case []int:
			if v[0] < len(e) {
				r = r.append(protect(v[1], e[v[0]]))
			}
		default:
			panic(fmt.Sprintf("unexpected template element: %v", reflect.TypeOf(t)))
		}
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
