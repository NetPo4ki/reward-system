ALTER TABLE user_tasks
    ADD CONSTRAINT user_tasks_user_id_task_id_unique
    UNIQUE (user_id, task_id);