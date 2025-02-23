class Parent inherits IO {
    print(): Object {
        out_string("LETS GO\n")
    };
};

class Main inherits IO {
    main(): Object {
        let p : Parent <- new Parent in p.print()
    };
};