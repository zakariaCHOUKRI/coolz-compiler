class Main inherits IO {
   main() : Object {{
      out_string("Please enter a number: ");
      let input_number : Int <- in_int()
      in {
         out_string("The number you inputted is: ");
         out_int(input_number);
         out_string("\n");
      };}
   };
};
