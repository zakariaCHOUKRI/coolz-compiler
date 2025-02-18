class Main inherits IO {
    main() : Object {
        {
            out_string("Enter your name: ");
            let name : String <- in_string() in {
                out_string("Hello, ").out_string(name).out_string("!\n");
                out_string("Enter your age: ");
                let age : Int <- in_int() in
                    out_string("You are ").out_int(age).out_string(" years old.\n");
            };
        }
    };
};
