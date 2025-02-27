class Main inherits IO{
    main(): Object {
        {
            out_string("Math module loaded.\n");
        }
    };
};

class GCD inherits IO{
    gcd(a: Int, b: Int) : Int{
        {
            if a = b then
                a
            else if a < b then 
                    gcd(b, a)
                 else gcd(b, a-b)
                 fi
            fi;
        }
    };
};
