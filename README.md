# goober-lisp

This is a program I wrote during a hackathon with the goal of learning Golang
by writing a simple lisp interpreter.

## getting started

You can run the following from the root of the repo to get the repl running.

```
$ ./repl.sh
```

To get a sense of what the simple language is capable of, take a look at the [core.el](core.el) where the builtins of the language are defined.

Here's the fibonacci sequence:

```lisp
user> (defn fib (n)
  (let (fib-n (fn (x n)
                (if (< (count x) n) 
                  (recur (cons (+ (first x) (second x)) x) n)
                  (reverse x))))
    (fib-n '(1 0) n)))

user> (fib 10)
> (0 1 1 2 3 5 8 13 21 34)
```
## Features

* value semantics
  ```lisp
  '((100 (200 (true))))
  ;; ((100 (200 (true))))
  ```
* dynamic bindings
  ```lisp
  (def x 100)
  (println 100)
  ;; 100
  ```
* lexical bindings
  ```lisp
  (let (x 100)
    (+ x 1))
  ;; 101
  ```
* macros
  ```lisp
  (defmacro when (test & rest) 
    (list 'if test (cons 'do rest)))

  (when (= (+ 1 1) 2) 'x)
  ;; 'x
  ```
* first-class functions (`fn`), tail recursion (`recur`)
  ```lisp
  (defn map (f coll)
    (let (map-inner (fn (old-coll new-coll)
                    (if (empty? old-coll)
                      new-coll
                      (recur
                        (rest old-coll)
                        (cons (f (first old-coll)) new-coll))))
        mapped (map-inner (seq coll) (list)))
    (reverse mapped)))
    
  (defn inc (n) (+ n 1))
    
  (map inc '(1 2 3))
  ;; (2 3 4)
  ```
## Status

Please don't use this interpreter for anything important, you will probably lose a finger. But I highly recommend Golang, and Lisp!
