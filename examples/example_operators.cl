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

                    let c : Int <- a / b in {
                        out_string("integer division of a/b: ");
                        out_int(c);
                    };
                };
            };
        }
    };
};