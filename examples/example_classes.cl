class Parent inherits IO {
    print(): Object {
        out_string("LETS GO\n")
    };
};

class Child inherits Parent {
};

class Main inherits IO {
    main(): Object {
        let c : Child <- new Child in c.print()
    };
};