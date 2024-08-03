# Ivycel

Ivycel is an experimental spreadsheet-like interface for [Ivy](https://github.com/robpike/ivy).

The initial motivation for the project was the desire for a spreadsheet
application that could handle calculations in binary and hexadecimal without
having to resort to cumbersome function calls.

When I discovered that Ivy handled binary and hexadecimal, I thought it would be
interesting to use it as the engine for such a spreadsheet. Ivycel is the result
of this experimentation

### Limitations

Currently, Ivycel offers the bare minimum of functionality that you might expect
from a spreadsheet application. The loading and saving of worksheets is currently
unimplemented.

The process of cell recalculation is not well thought out and certainly requires
optimisation. There may also be situations where the recalcuation will be incomplete.

The interface with Ivy is entirely through Ivy's run.Run() function. Ivy has not been
changed at all.

All Ivy expressions should work except the "special commands" and the `op` and
`opdelete` commands. 

The GUI is created using [GIU](https://github.com/AllenDang/giu) which itself is based on [Dear Imgui](https://github.com/ocornut/imgui). Unfortunatly, GIU has additional development requirements see [the 'install' instructions for GIU](https://github.com/AllenDang/giu?tab=readme-ov-file#install) for more information.

Currently, the GUI code is entirely in the main.go file and definitely
requires tidying and clarifying.

### Demonstration Videos

The following videos demonstrate Ivycel in operation, showing how a spreadsheet
style interface can work with Ivy.

1) The first video shows basic arithmetic and referencing of cells. References can be typed
   or the referenced cell can be double-clicked. Double-clicking is often more convenient
   because the cell reference must be inside braces.
  
   The ability of Ivy to output results in different number bases is also demonstrated.

https://github.com/user-attachments/assets/a3099eb5-6cd2-45a6-992f-2994854610c6

2) Some Ivy results come in parts. In this video the Ivy keyword `iota` is used.

   The video also shows how a reference to a cell can be made using the context menu (accessible
   with the right mouse button on most computers).

https://github.com/user-attachments/assets/502e2d3f-3764-4f12-8c68-b61c960a491d

3) This video demonstrates how the results of the Ivy keyword `rho` are displayed. Similar to the
   previous example, individual elements in the result can be references or the expression as a
   whole can be referenced.

https://github.com/user-attachments/assets/d78bf7b3-2aec-4807-93a7-ade3ac7af90f

4) A much larger `rho` result in this video. Also, the video shows how a multi-part result will be
   displayed when it comes into contact with another result (ie. an occupied cell).

   The cropped result is still valid and can be referenced but requires the index notation.

   Deleting the blocking expression, causes the `iota` expression to be re-evaluated and the result shown in full.

https://github.com/user-attachments/assets/e47c698b-802d-433d-b9d5-842e9a07a717

5) This video shows the insertion of columns and rows. In particular, it shows how references are preserved. It also
   shows the effect insertion has on multi-part results.

   For example, when the column is inserted at column F, the `iota` result is moved along but the `rho` result is
   unaffected (because the `rho` expression is rooted at a cell before column F.
   
https://github.com/user-attachments/assets/7f0c8412-9c7e-44e2-b4f9-1dce5bbf5f82

6) This final video demonstrates some features of Ivy. Rational numbers, exponents, trignonometry and complex numbers.

https://github.com/user-attachments/assets/175f4a06-a9ee-43cb-93ee-ce260c73e515






