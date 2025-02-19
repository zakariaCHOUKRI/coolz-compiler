class Main inherits IO {
   main() : Object {
      {
         out_string("Enter a string: ");
         out_string(in_string());
         out_string("\n");
         
         out_string("Enter a number: ");
         out_int(in_int());
         out_string("\n");
      }
   };
};