(* COOL Program Demonstrating Method Overriding *)

class Animal inherits IO{
    speak() : String {
        "I am an animal.\n"
    };
};

class Dog inherits Animal {
    speak() : String {
        "Woof!\n"
    };
};

class Cat inherits Animal {
    speak() : String {
        "Meow!\n"
    };
};


class Bird inherits Dog {
    speak() : String {
        "Chirp!\n"
    };
};

class Fish inherits Cat {
    speak() : String {
        "Blub blub!\n"
    };
};

class Parrot inherits Bird {
    speak() : String {
        "Polly wants a cracker!\n"
    };
};

class Goldfish inherits Fish {
    speak() : String {
        "Glub glub!\n"
    };
};


class Main inherits IO{
    main() : Object {
        let animal : Animal <- new Animal,
            dog : Dog <- new Dog,
            cat : Cat <- new Cat,
            bird : Bird <- new Bird
        in {
            out_string(animal.speak());
            out_string(dog.speak());
            out_string(cat.speak());
            out_string(bird.speak());
        }
    };
};