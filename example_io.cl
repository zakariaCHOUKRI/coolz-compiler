class Main inherits IO {
    main() : Object {
        {
            out_string("Hello. Please enter a number: ");
            let num : Int <- in_int() in {
                out_string("You entered: ").out_int(num).out_string("\n");
            };
        }
    };
};
