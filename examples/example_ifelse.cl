class Main inherits IO {
    main() : Object {
        {
            if 2 < 1 then
                out_string("true\n")
            else
                out_string("false\n")
            fi;

            if 0 < 1 then
                out_string("true\n")
            else
                out_string("false\n")
            fi;
        }
    };
};