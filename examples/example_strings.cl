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
                println("testing length() ...");
                println("the length of:");
                print("\"");
                print(x);
                print("\" ");
                print("is: ");
                out_int(x.length());
                println("");
                println("");
            };

            let y : String <- "0123456789" in {
                println("testing substr() ...");
                println("substring index from 3 to 7 of the string \"0123456789\" is:");
                println(y.substr(3, 7-3+1));
                println("");
            };

            print("Please input a string a: ");
            let a : String <- in_string() in {
                print("Please input a string b: ");
                let b : String <- in_string() in {
                    let c : String <- a.concat(b) in {
                        println("The concatenation of a and b is: ");
                        println(c);
                    };
                };
            };
        }
    };
};