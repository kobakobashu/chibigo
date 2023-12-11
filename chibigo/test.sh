#!/bin/bash
assert() {
  expected="$1"
  input="$2"

  ./chibigo "$input" > tmp.s
  cc -o tmp tmp.s
  ./tmp
  actual="$?"

  if [ "$actual" = "$expected" ]; then
    echo "$input => $actual"
  else
    echo "$input => $expected expected, but got $actual"
    exit 1
  fi
}

assert 0 '{ return 0; }'
assert 42 '{ return 42; }'
assert 21 '{ return 5+20-4; }'
assert 41 '{ return  12 + 34 - 5 ; }'
assert 47 '{ return 5+6*7; }'
assert 15 '{ return 5*(9-6); }'
assert 4 '{ return (3+5)/2; }'
assert 10 '{ return -10+20; }'
assert 10 '{ return - -10; }'
assert 10 '{ return - - +10; }'

assert 0 '{ return 0==1; }'
assert 1 '{ return 42==42; }'
assert 1 '{ return 0!=1; }'
assert 0 '{ return 42!=42; }'

assert 1 '{ return 0<1; }'
assert 0 '{ return 1<1; }'
assert 0 '{ return 2<1; }'
assert 1 '{ return 0<=1; }'
assert 1 '{ return 1<=1; }'
assert 0 '{ return 2<=1; }'

assert 1 '{ return 1>0; }'
assert 0 '{ return 1>1; }'
assert 0 '{ return 1>2; }'
assert 1 '{ return 1>=0; }'
assert 1 '{ return 1>=1; }'
assert 0 '{ return 1>=2; }'

assert 3 '{ var a int; a=3; return a; }'
assert 3 '{ var a int=3; return a; }'
assert 8 '{ var a int=3; var z int=5; return a+z; }'

assert 6 '{ var a int; var b int; a=b=3; return a+b; }'
assert 3 '{ var foo int=3; return foo; }'
assert 8 '{ var foo123 int=3; var bar int=5; return foo123+bar; }'

assert 1 '{ return 1; 2; 3; }'
assert 2 '{ 1; return 2; 3; }'
assert 3 '{ 1; 2; return 3; }'

assert 3 '{ {1; {2;} return 3;} }'

assert 5 '{ ;;; return 5; }'

assert 3 '{ if 1 == 0 { return 2; }; return 3; }'
assert 2 '{ if 1 == 1 { return 2; }; return 3; }'
assert 2 '{ if 1-1 == 0 { return 2; }; return 3; }'
assert 4 '{ if 1 == 0 { 1; 2; return 3; } else { return 4; } }'
assert 3 '{ if 1 == 1 { 1; 2; return 3; } else { return 4; } }'

assert 55 '{ var i int=0; var j int=0; for i=0; i<=10; i=i+1 { j=i+j; }; return j; }'
assert 3 '{ for { return 3; } return 5; }'

assert 10 '{ var i int=0; for i<10 { i=i+1; }; return i; }'

assert 3 '{ var x int=3; return *&x; }'
assert 3 '{ var x int=3; var y *int=&x; var z **int=&y; return **z; }'
assert 5 '{ var x int=3; var y *int=&x; *y=5; return x; }'

echo OK