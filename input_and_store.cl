class Main inherits IO {
   main() : Object {
      {
         out_string("Please enter your name: ");
         let input_string : String <- in_string()
         in {
            out_string("Your name is: ");
            out_string(input_string);
            out_string("\n");
         };

         out_string("Please enter a number: ");
         let input_number2 : Int <- in_int()
         in {
            out_string("Your number is: ");
            out_int(input_number2);
            out_string("\n");
         };

         out_string("Please enter a number: ");
         let input_number : Int <- in_int()
         in {
            out_string("Your number is: ");
            out_int(input_number);
            out_string("\n");
         };

         out_string("Please enter your name: ");
         let input_string2 : String <- in_string()
         in {
            out_string("Your name is: ");
            out_string(input_string2);
            out_string("\n");
         };
      }
   };
};
