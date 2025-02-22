class Main inherits IO {
    main() : Object {
        {
            let a : Int <- 15 in {
                let b : Int <- 4 in {
                    out_string("a: ");
                    out_int(a);
                    out_string("\n");

                    out_string("b: ");
                    out_int(b);
                    out_string("\n");


                    out_string("a / b = ");
                    out_int(a/b);
                    out_string(" (integer division)\n");

                    out_string("a + b = ");
                    out_int(a+b);
                    out_string("\n");

                    out_string("a * b = ");
                    out_int(a*b);
                    out_string("\n");

                    out_string("b - a = ");
                    out_int(b-a);
                    out_string("\n");

                    out_string("testing pratt parsing for\n");
                    out_string("1 - 2 + 3 * 4 + a * b - (2 + 3) = ");
                    out_int(1 - 2 + 3 * 4 + a*b - (2+3));
                };
            };
        }
    };
};