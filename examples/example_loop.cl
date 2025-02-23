class Main inherits IO {
    main() : Object {
        {
            out_string("Testing loops:\n");
            let i : Int <- 0
            in {
                while (i <= 10) loop {
                    out_string("i: ");
                    out_int(i);
                    out_string("\n");
                    i <- i + 1;
                }
                pool;
            };

            out_string("Testing loops:\n");
            let j : Int <- 10
            in {
                while (0 <= j) loop {
                    out_string("j: ");
                    out_int(j);
                    out_string("\n");
                    j <- j - 1;
                }
                pool;
            };
        }
    };
};