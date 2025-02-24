class Parent inherits IO {
    print(x: String): Object {
        {
            out_string(x);
            out_string("\n");
        }
    };
};

class Child inherits Parent {
};

class Main inherits IO {
    main(): Object {
        let c : Child <- new Child in c.print("Let's go");
    };
};