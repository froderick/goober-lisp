(defmacro defn (name args & rest)
  (let (f (cons 'fn (cons args rest)))
   (list 'def name f)))

(defn not (x) (if x false true))

(defmacro when (test & rest) 
  (list 'if test (cons 'do rest)))

(defn empty? (coll) (= (count coll) 0))

(defn inc (n) (+ n 1))

(defn reverse (coll)
  (let (map-inner (fn (old-coll new-coll)
                    (if (empty? old-coll)
                      new-coll
                      (recur
                        (rest old-coll)
                        (cons (first old-coll) new-coll)))))
    (map-inner coll (list))))

(defn map (f coll)
  (let (map-inner (fn (old-coll new-coll)
                    (if (empty? old-coll)
                      new-coll
                      (recur
                        (rest old-coll)
                        (cons (f (first old-coll)) new-coll))))
        mapped (map-inner (seq coll) (list)))
    (reverse mapped)))

