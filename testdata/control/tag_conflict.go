package control

/* @openapi
tags:
  - name: tasks
    description: One description
*/

func TagConflictA() {}

/* @openapi
tags:
  - name: tasks
    description: Another description
*/

func TagConflictB() {}
