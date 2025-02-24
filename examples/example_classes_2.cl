class Parent inherits IO {
    print(x: String): Object {
        {
            out_string(x);
            out_string("\n");
        }
    };
};

class Child inherits Parent {
    print2(x: String, y: String) : Object {
        {
            print(x);
            print(y);
        }
    };
};

class Main inherits Child{
    main(): Object {
        {
            let c : Child <- new Child in {
                c.print("Let's go");
                c.print(c.type_name());
            };

            print2("1", "2");
            self.print("this should print");
            abort();
            print("this should not print");
        }
    };
};