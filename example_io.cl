class Main inherits IO {
    main() : Object {
        {
            out_string("Hello.\nPlease enter a string: ");
            flush();
            let user_input : String <- in_string() in {
                out_string("You entered: ").out_string(user_input).out_string("\n");
            };
        }
    };
};
