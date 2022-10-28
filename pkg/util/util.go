package util

func Contains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}

func Clean(s []string, e string) ([]string, bool) {

   c := false

   // Find and remove "two"
   for i, v := range s {
      if v == e {
         s = append(s[:i], s[i+1:]...)
         c = true
         break
      }
   }

   return s, c
}
