#!/bin/bash
cat <<EOF | gcc -xc -c -o tmp2.o -
int ret3() { return 3; }
int ret5() { return 5; }
int add(int x, int y) { return x+y; }
int sub(int x, int y) { return x-y; }
int add6(int a, int b, int c, int d, int e, int f) {
  return a+b+c+d+e+f;
}
EOF

assert() {
  expected="$1"
  input="$2"

  echo "$input" | ./chibigo - > tmp.s || exit
  cc -o tmp tmp.s tmp2.o
  ./tmp
  actual="$?"

  if [ "$actual" = "$expected" ]; then
    echo "$input => $actual"
  else
    echo "$input => $expected expected, but got $actual"
    exit 1
  fi
}

assert 0 'func main() int { return 0; }'
assert 42 'func main() int { return 42; }'
assert 21 'func main() int { return 5+20-4; }'
assert 41 'func main() int { return  12 + 34 - 5 ; }'
assert 47 'func main() int { return 5+6*7; }'
assert 15 'func main() int { return 5*(9-6); }'
assert 4 'func main() int { return (3+5)/2; }'
assert 10 'func main() int { return -10+20; }'
assert 10 'func main() int { return - -10; }'
assert 10 'func main() int { return - - +10; }'

assert 0 'func main() int { return 0==1; }'
assert 1 'func main() int { return 42==42; }'
assert 1 'func main() int { return 0!=1; }'
assert 0 'func main() int { return 42!=42; }'

assert 1 'func main() int { return 0<1; }'
assert 0 'func main() int { return 1<1; }'
assert 0 'func main() int { return 2<1; }'
assert 1 'func main() int { return 0<=1; }'
assert 1 'func main() int { return 1<=1; }'
assert 0 'func main() int { return 2<=1; }'

assert 1 'func main() int { return 1>0; }'
assert 0 'func main() int { return 1>1; }'
assert 0 'func main() int { return 1>2; }'
assert 1 'func main() int { return 1>=0; }'
assert 1 'func main() int { return 1>=1; }'
assert 0 'func main() int { return 1>=2; }'

assert 3 'func main() int { var a int; a=3; return a; }'
assert 3 'func main() int { var a int=3; return a; }'
assert 8 'func main() int { var a int=3; var z int=5; return a+z; }'

assert 6 'func main() int { var a int; var b int; a=b=3; return a+b; }'
assert 3 'func main() int { var foo int=3; return foo; }'
assert 8 'func main() int { var foo123 int=3; var bar int=5; return foo123+bar; }'

assert 1 'func main() int { return 1; 2; 3; }'
assert 2 'func main() int { 1; return 2; 3; }'
assert 3 'func main() int { 1; 2; return 3; }'

assert 3 'func main() int { {1; {2;} return 3;} }'

assert 5 'func main() int { ;;; return 5; }'

assert 3 'func main() int { if 1 == 0 { return 2; }; return 3; }'
assert 2 'func main() int { if 1 == 1 { return 2; }; return 3; }'
assert 2 'func main() int { if 1-1 == 0 { return 2; }; return 3; }'
assert 4 'func main() int { if 1 == 0 { 1; 2; return 3; } else { return 4; } }'
assert 3 'func main() int { if 1 == 1 { 1; 2; return 3; } else { return 4; } }'

assert 55 'func main() int { var i int=0; var j int=0; for i=0; i<=10; i=i+1 { j=i+j; }; return j; }'
assert 3 'func main() int { for { return 3; } return 5; }'

assert 10 'func main() int { var i int=0; for i<10 { i=i+1; }; return i; }'

assert 3 'func main() int { var x int=3; return *&x; }'
assert 3 'func main() int { var x int=3; var y *int=&x; var z **int=&y; return **z; }'
assert 5 'func main() int { var x int=3; var y *int=&x; *y=5; return x; }'

assert 8 'func main() int { var a, b int; a=3; b=5; return a+b; }'
assert 8 'func main() int { var a, b int=3, 5; return a+b; }'

assert 3 'func main() int { return ret3(); }'
assert 5 'func main() int { return ret5(); }'

assert 8 'func main() int { return add(3, 5); }'
assert 2 'func main() int { return sub(5, 3); }'
assert 21 'func main() int { return add6(1,2,3,4,5,6); }'
assert 66 'func main() int { return add6(1,2,add6(3,4,5,6,7,8),9,10,11); }'
assert 136 'func main() int { return add6(1,2,add6(3,add6(4,5,6,7,8,9),10,11,12,13),14,15,16); }'

assert 32 'func main() int { return ret32(); } func ret32() int { return 32; }'

assert 7 'func main() int { return add2(3,4); } func add2(x int, y int) int { return x+y; }'
assert 1 'func main() int { return sub2(4,3); } func sub2(x int, y int) int { return x-y; }'
assert 55 'func main() int { return fib(9); } func fib(x int) int { if (x<=1) return 1; return fib(x-1) + fib(x-2); }'

assert 1 'func main() int { var x [2]int; x[0] = 1; return x[0]; }'
assert 2 'func main() int { var x [2]int; x[1] = 2; return x[1]; }'
assert 1 'func main() int { var x [2]*int; var y int = 1; x[0] = &y; return *x[0]; }'
assert 2 'func main() int { var x [2]*int; var y int = 2; x[1] = &y; return *x[1]; }'

assert 2 'func main() int { var x [2][3]int; x[0][0] = 2; return x[0][0]; }'
assert 3 'func main() int { var x [2][3]int; x[1][1] = 3; return x[1][1]; }'

assert 0 'var x int; func main() int { return x; }'
assert 3 'var x int; func main() int { x=3; return x; }'
assert 7 'var x int; var y int; func main() int { x=3; y=4; return x+y; }'
assert 7 'var x, y int; func main() int { x=3; y=4; return x+y; }'
assert 0 'var x [4]int; func main() int { x[0]=0; x[1]=1; x[2]=2; x[3]=3; return x[0]; }'
assert 1 'var x [4]int; func main() int { x[0]=0; x[1]=1; x[2]=2; x[3]=3; return x[1]; }'
assert 2 'var x [4]int; func main() int { x[0]=0; x[1]=1; x[2]=2; x[3]=3; return x[2]; }'
assert 3 'var x [4]int; func main() int { x[0]=0; x[1]=1; x[2]=2; x[3]=3; return x[3]; }'

assert 1 'func main() char { var x char=1; return x; }'
assert 1 'func main() char { var x char=1; var y char=2; return x; }'
assert 2 'func main() char { var x char=1; var y char=2; return y; }'

assert 1 'func main() char { return subChar(7, 3, 3); } func subChar(a char, b char, c char) int { return a-b-c; }'

assert 0 'func main() int { return ""[0]; }'

assert 97 'func main() int { return "abc"[0]; }'
assert 98 'func main() int { return "abc"[1]; }'
assert 99 'func main() int { return "abc"[2]; }'
assert 0 'func main() int { return "abc"[3]; }'

assert 2 'func main() int { /* return 1; */ return 2; }'
assert 2 'func main() int { // return 1;
return 2; }'

assert 2 'func main() int { var x int=2; { var x int=3; } return x; }'
assert 2 'func main() int { var x int=2; { var x int=3; } { var y int=4; return x; }}'
assert 3 'func main() int { var x int=2; { x=3; } return x; }'

assert 3 'var x int = 3; func main() int { return x;}'
assert 2 'var x int = 1; func main() int { x=2; return x;}'
assert 2 'var x int = 1; func main() int { {x = 2;} return x;}'
assert 2 'var x [3]int; func main() int { x[1] = 2; return x[1];}'

echo OK