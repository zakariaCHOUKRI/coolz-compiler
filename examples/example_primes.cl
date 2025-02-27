class Main inherits IO {
   findPrimes(max : Int) : Object {
      let i : Int <- 2,
          j : Int <- 2,
          isPrime : Bool in
      {
         while i <= max loop
         {
            isPrime <- true;
            j <- 2;
            
            while j * j <= i loop
            {
               if i - (i/j)*j = 0 then
               {
                  isPrime <- false;
                  j <- i;  (* Break inner loop *)
               }
               else
                  j <- j + 1
               fi;
            } pool;
            
            if isPrime then
            {
               out_int(i);
               out_string(" ");
            }
            else 0
            fi;
            
            i <- i + 1;
         } pool;
         out_string("\n");
      }
   };

   main() : Object {
      {
         out_string("Prime numbers up to 30: ");
         findPrimes(30);
      }
   };
};