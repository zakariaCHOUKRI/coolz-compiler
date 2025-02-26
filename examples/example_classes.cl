class Parent inherits IO {
    print(x: String): Object {
        {
            out_string(x);
            out_string("\n");
        }
    };

    polymorphism() : Object {
        {
            out_string("parent");
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

    polymorphism() : Object {
        {
            out_string("child");
            out_string("\n");
        }
    };
};

class Main inherits Child{
    main(): Object {
        {
            let c : Child <- new Child in {
                c.print("Let's go");
                c.print(c.type_name());
                c.polymorphism();
            };

            let p : Parent <- new Parent in {
                p.polymorphism();
            };

            print2("1", "2");
            self.print("this should be printed");
            print(type_name());

            abort();
            print("this should not be printed");
        }
    };
};