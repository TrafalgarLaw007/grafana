import React, { FC } from 'react';
import { Button, Tooltip, Icon, Form, Input, Field, FieldSet } from '@grafana/ui';
import { User } from 'app/types';
import config from 'app/core/config';
import { ProfileUpdateFields } from 'app/core/utils/UserProvider';

export interface Props {
  user: User;
  isSavingUser: boolean;
  updateProfile: (payload: ProfileUpdateFields) => void;
}

const { disableLoginForm } = config;

export const UserProfileEditForm: FC<Props> = ({ user, isSavingUser, updateProfile }) => {
  const onSubmitProfileUpdate = (data: ProfileUpdateFields) => {
    updateProfile(data);
  };

  return (
    <Form onSubmit={onSubmitProfileUpdate} validateOn="onBlur">
      {({ register, errors }) => {
        return (
          <FieldSet label="编辑个人资料">
            <Field label="姓名" invalid={!!errors.name} error="Name is required">
              <Input name="name" ref={register({ required: true })} placeholder="姓名" defaultValue={user.name} />
            </Field>
            <Field label="邮箱" invalid={!!errors.email} error="Email is required" disabled={disableLoginForm}>
              <Input
                name="email"
                ref={register({ required: true })}
                placeholder="邮箱"
                defaultValue={user.email}
                suffix={<InputSuffix />}
              />
            </Field>
            <Field label="用户名" disabled={disableLoginForm}>
              <Input
                name="login"
                ref={register}
                defaultValue={user.login}
                placeholder="用户名"
                suffix={<InputSuffix />}
              />
            </Field>
            <div className="gf-form-button-row">
              <Button variant="primary" disabled={isSavingUser}>
                Save
              </Button>
            </div>
          </FieldSet>
        );
      }}
    </Form>
  );
};

export default UserProfileEditForm;

const InputSuffix: FC = () => {
  return disableLoginForm ? (
    <Tooltip content="Login Details Locked - managed in another system.">
      <Icon name="lock" />
    </Tooltip>
  ) : null;
};
