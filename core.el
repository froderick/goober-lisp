(def not (fn (x) (if x false true)))

(def empty? (fn (coll) (= (count coll) 0)))

(def inc (fn (n) (+ n 1)))

(def reverse 
  (fn (coll)
    (let (map-inner (fn (old-coll new-coll)
                      (if (empty? old-coll)
                        new-coll
                        (recur
                          (rest old-coll)
                          (cons (first old-coll) new-coll)))))
      (map-inner coll (list)))))

(def map 
  (fn (f coll)
    (let (map-inner (fn (old-coll new-coll)
                      (if (empty? old-coll)
                        new-coll
                        (recur
                          (rest old-coll)
                          (cons (f (first old-coll)) new-coll))))
          mapped (map-inner (seq coll) (list)))
      (reverse mapped))))

