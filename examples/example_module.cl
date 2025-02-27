import math;

class Main inherits IO{
    main(): Object {
        let solver: GCD <- new GCD
        in {
            out_string("gcd(15, 35) = ");
            out_int(solver.gcd(15, 35));
            out_string("\n");

            out_string("gcd(7, 49) = ");
            out_int(solver.gcd(7, 49));
            out_string("\n");

            out_string("gcd(7, 39) = ");
            out_int(solver.gcd(7, 39));
            out_string("\n");
        }
    };
};
