
export const authVerifier = {
  accessUser: ({ user }) => {
    if (!user) {
      throw new Error('NotAuthorized')
    }
  },
  accessAdmin: ({ user }) => {
    if (!user) {
      throw new Error('NotAuthorized')
    }
  }
}
