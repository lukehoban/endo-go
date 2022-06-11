package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
)

type DNA string

func (d DNA) get(i int) byte {
	if i >= len(d) {
		return 0
	}
	return d[i]
}

func (d DNA) skip(i int) DNA {
	return d[i:]
}

func (d DNA) append(d2 DNA) DNA {
	return d + d2
}

func emptyDNA() DNA {
	return DNA("")
}

func (d DNA) substring(i int, end int) DNA {
	if end > len(d) {
		end = len(d)
	}
	return d[i:end]
}

func (d DNA) String() string {
	snip := d.substring(0, 10)
	cont := ""
	if len(d) > 10 {
		cont = "..."
	}
	return fmt.Sprintf("%s%s (%d bases)", string(snip), cont, len(d))
}

var dna DNA
var rna []DNA

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

	dna = DNA(prefix).append(DNA(byts))

	iteration := 0
	for {

		fmt.Printf("\niteration %d\n", iteration)
		fmt.Printf("dna = %s\n", dna)

		pat, err := pattern()
		if err != nil {
			return err
		}
		fmt.Printf("pat = %s\n", pat)

		tmpl, err := template()
		if err != nil {
			return err
		}
		fmt.Printf("tmpl = %s\n", tmpl)

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
	var e []DNA
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
			if i > len(dna) {
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
func findPostfix(d DNA, needle string) int {
outer:
	for i := 0; i <= len(d)-len(needle); i++ {
		for j := 0; j < len(needle); j++ {
			if d.get(i+j) != needle[j] {
				continue outer
			}
		}
		return i + len(needle)
	}
	return -1
}

func replace(tmpl Template, e []DNA) {
	r := emptyDNA()
	for _, t := range tmpl {
		switch v := t.(type) {
		case int32:
			r = r.append(DNA(string(v)))
		case int:
			x := 0
			if v < len(e) {
				x = len(e[v])
			}
			r = r.append(asnat(x))
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
func asnat(i int) DNA {
	ret := ""
	for {
		if i == 0 {
			ret += "P"
			return DNA(ret)
		}
		if i%2 == 0 {
			ret += "I"
		} else {
			ret += "C"
		}
		i /= 2
	}
}

func protect(l int, d DNA) DNA {
	for i := 0; i < l; i++ {
		d = quote(d)
	}
	return d
}

// ICFPI => CFPICC
func quote(d DNA) DNA {
	ret := emptyDNA()
	for i := 0; i < len(d); i++ {
		switch d.get(i) {
		case 'I':
			ret = ret.append("C")
		case 'C':
			ret = ret.append("F")
		case 'F':
			ret = ret.append("P")
		case 'P':
			ret = ret.append("IC")
		default:
			panic("invalid base")
		}
	}
	return ret
}

func main() {
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
