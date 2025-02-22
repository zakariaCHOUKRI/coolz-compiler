class Main inherits IO {
    main() : Object {
        {
            if 2 < 1 then
                out_string("true\n")
            else
                out_string("false\n")
            fi;

            out_string("Please enter your name: ");
            let input_string : String <- in_string()
            in {
                out_string("Your name is: ");
                out_string(input_string);
                out_string("\n");
            };

            if 1 < 2 then
                out_string("true\n")
            else
                out_string("false\n")
            fi;
        }
    };
};