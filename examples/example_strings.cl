class Printer inherits IO {
    print(x: String) : Object {
        {
            out_string(x);
        }
    };
    
    println(x: String) : Object {
        {
            print(x);
            print("\n");
        }
    };
};

class Main inherits Printer {
    main(): Object {
        {
            let x : String <- "Hello everyone" in {
                println("the length of:");
                print("\"");
                print(x);
                print("\" ");
                print("is: ");
                out_int(x.length());
            };
        }
    };
};