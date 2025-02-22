class Main inherits IO {
    main() : Object {
        {
            out_string("Testing loops:\n");
            let i : Int <- 0
            in {
                while i < 10
                loop {
                    out_string("i: ");
                    out_int(i);
                    out_string("\n");
                    i <- i + 1;
                } pool;;
            };
        }
    };
};