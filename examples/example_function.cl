class Main inherits IO {
    main(): Object {
        {
            self.hello();
            let x : Int <- 6 in {
                out_string("6! = ");
                out_int(factorial(x));
            };
        }
    };

    factorial(n: Int): Int {
        {
            if n = 0 then {
                1;
            } else {
                n * factorial(n - 1);
            } fi;
        }
    };

    hello(): Object {
        {
            out_string("Hello user :)\n");
        }
    };
};