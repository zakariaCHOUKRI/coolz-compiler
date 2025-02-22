class Main inherits IO {
    main(): Object {
        {
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
};